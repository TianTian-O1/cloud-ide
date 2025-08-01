package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	mv1 "github.com/mangohow/cloud-ide/cmd/control-plane/internal/api/v1"
	"github.com/mangohow/cloud-ide/pkg/notifier"
	"github.com/mangohow/cloud-ide/pkg/pb"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkSpaceService struct {
	pb.UnimplementedCloudIdeServiceServer
	logger    logr.Logger
	client    client.Client
	waiter    notifier.Waiter
	namespace string
}

func NewWorkSpaceService(c client.Client, logger logr.Logger, waiter notifier.Waiter, namespace string) *WorkSpaceService {
	return &WorkSpaceService{
		logger:    logger,
		client:    c,
		waiter:    waiter,
		namespace: namespace,
	}
}

var _ = pb.CloudIdeServiceServer(&WorkSpaceService{})

const (
	WorkspaceAlreadyExist = "workspace already exist"
	WorkspaceNotExist     = "workspace not exist"
	WorkspaceCreateFailed = "create workspace error"
	WorkspaceStartFailed  = "start workspace error"
	WorkspaceStopFailed   = "stop workspace error"
	WorkspaceDeleteFailed = "delete workspace error"
)

const WorkspaceNameFormat = "ws-%s-%s"

// CreateSpace 创建并且启动Workspace,将Operation字段置为"Start",当Workspace被创建时,PVC和Pod也会被创建
// 该接口仅被用于第一次创建工作空间并且启动
func (s *WorkSpaceService) CreateSpace(ctx context.Context, info *pb.RequestCreate) (*pb.ResponseCreate, error) {
	var (
		ws  = &mv1.WorkSpace{}
		res = &pb.ResponseCreate{}
	)

	// 校验参数
	err := s.validateRequestCreate(info)
	if err != nil {
		s.logger.Error(err, "request param invalid")
		res.Status = pb.ResponseCreate_Error
		return res, status.Error(codes.InvalidArgument, err.Error())
	}

	// 1.先查询workspace是否存在
	name := workspaceName(info.Uid, info.Sid)
	exist := s.checkWorkspaceExist(ctx, client.ObjectKey{Name: name, Namespace: s.namespace}, ws)
	stus := status.New(codes.AlreadyExists, WorkspaceAlreadyExist)
	if exist {
		res.Status = pb.ResponseCreate_AlreadyExist
		res.Message = WorkspaceAlreadyExist
		return res, stus.Err()
	}

	// 2.如果不存在就创建
	w := s.constructWorkspace(info, name)
	if err := s.client.Create(ctx, w); err != nil {
		if errors.IsAlreadyExists(err) {
			res.Status = pb.ResponseCreate_AlreadyExist
			res.Message = WorkspaceAlreadyExist
			return res, stus.Err()
		}

		s.logger.Error(err, "create workspace")
		res.Status = pb.ResponseCreate_Error
		res.Message = WorkspaceCreateFailed
		return res, status.Error(codes.Unknown, err.Error())
	}

	// 3.等待Pod处于Running状态
	err = s.waitForPodRunning(ctx, client.ObjectKey{Name: w.Name, Namespace: w.Namespace}, w)
	if err != nil {
		s.logger.Error(err, "wait for pod running")
		res.Status = pb.ResponseCreate_Error
		res.Message = WorkspaceStartFailed
		return res, status.Error(codes.ResourceExhausted, WorkspaceStartFailed)
	}

	return res, nil
}

func (s *WorkSpaceService) waitForPodRunning(ctx context.Context, key client.ObjectKey, ws *mv1.WorkSpace) error {
	s.logger.Info("waiting for pod ready", "name", key.Name)

	c, cancelFunc := context.WithTimeout(ctx, time.Second*300)
	defer cancelFunc()
	// 1.等待Pod可用，最久等待300s (5分钟) - 增加超时时间以适应高负载环境
	err := s.waiter.WaitFor(c, ws.Spec.SID)
	if err == nil {
		return nil
	}

	s.logger.Error(err, "wait for pod ready")
	// 2.处理错误情况,停止工作空间
	_, err = s.StopSpace(ctx, &pb.RequestStop{
		Uid: ws.Spec.UID,
		Sid: ws.Spec.SID,
	})
	if err != nil {
		s.logger.Error(err, "stop workspace")
	}

	return err
}

// StartSpace 启动Workspace
func (s *WorkSpaceService) StartSpace(ctx context.Context, req *pb.RequestStart) (*pb.ResponseStart, error) {
	if err := s.validateResourceLimit(req.ResourceLimit); err != nil {
		s.logger.Error(err, "request param invalid")
		return &pb.ResponseStart{}, err
	}

	res := &pb.ResponseStart{}

	// 1.先获取workspace,如果不存在返回错误
	var ws mv1.WorkSpace
	key := client.ObjectKey{
		Name:      workspaceName(req.Uid, req.Sid),
		Namespace: s.namespace,
	}
	exist := s.checkWorkspaceExist(ctx, key, &ws)
	if !exist {
		res.Status = pb.ResponseStart_NotFound
		res.Message = WorkspaceNotExist
		return res, status.Error(codes.NotFound, WorkspaceNotExist)
	}

	// 2.判断workspace是否处于运行或启动中状态
	if ws.Status.Phase == mv1.WorkspacePhaseStaring || ws.Status.Phase == mv1.WorkspacePhaseRunning {
		return res, nil
	}

	// 3.Pod的配置可能会改变
	ws.Spec.Cpu = req.ResourceLimit.Cpu
	ws.Spec.Memory = req.ResourceLimit.Memory
	// TODO storage改变需要特殊处理
	// ws.Spec.Storage = req.ResourceLimit.Storage

	// 4.更新Workspace的Operation字段以启动,使用RetryOnConflict,当资源版本冲突时重试
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 每次更新前要获取最新的版本
		var p mv1.WorkSpace
		exist := s.checkWorkspaceExist(ctx, key, &p)
		if !exist {
			return nil
		}

		// 更新workspace的Operation字段
		ws.Spec.Command = mv1.WorkSpaceStart
		if err := s.client.Update(ctx, &ws); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error(err, "update workspace")
		res.Status = pb.ResponseStart_Error
		res.Message = WorkspaceStartFailed
		return res, status.Error(codes.Unknown, err.Error())
	}

	if !exist {
		res.Status = pb.ResponseStart_NotFound
		res.Message = WorkspaceNotExist
		return res, status.Error(codes.NotFound, WorkspaceNotExist)
	}

	err = s.waitForPodRunning(ctx, key, &ws)
	if err != nil {
		s.logger.Error(err, "wait for pod running")
		res.Status = pb.ResponseStart_Error
		res.Message = WorkspaceStartFailed
		return res, status.Error(codes.ResourceExhausted, WorkspaceStartFailed)
	}

	return res, nil
}

// DeleteSpace 只需要将workspace删除即可,controller会负责删除对应的Pod和PVC
func (s *WorkSpaceService) DeleteSpace(ctx context.Context, req *pb.RequestDelete) (*pb.ResponseDelete, error) {
	res := &pb.ResponseDelete{}
	// 先查询是否存在,如果不存在则无需删除
	var ws mv1.WorkSpace
	name := workspaceName(req.Uid, req.Sid)
	exist := s.checkWorkspaceExist(ctx, client.ObjectKey{Name: name, Namespace: s.namespace}, &ws)
	if !exist {
		return res, nil
	}

	// 删除Workspace
	if err := s.client.Delete(ctx, &ws); err != nil {
		s.logger.Error(err, "delete workspace")
		res.Status = pb.ResponseDelete_Error
		res.Message = WorkspaceDeleteFailed
		return res, status.Error(codes.Unknown, err.Error())
	}

	return res, nil
}

// StopSpace 停止Workspace,只需要删除对应的Pod,因此修改Workspace的操作为Stop即可
func (s *WorkSpaceService) StopSpace(ctx context.Context, req *pb.RequestStop) (*pb.ResponseStop, error) {
	res := &pb.ResponseStop{}

	// 1.先查询Workspace是否存在，不存在则直接返回
	var ws mv1.WorkSpace
	name := workspaceName(req.Uid, req.Sid)
	err := s.client.Get(ctx, client.ObjectKey{Name: name, Namespace: s.namespace}, &ws)
	if errors.IsNotFound(err) {
		res.Status = pb.ResponseStop_NotFound
		res.Message = WorkspaceNotExist
		return res, status.Error(codes.NotFound, WorkspaceNotExist)
	}

	// 2.工作空间正在停止或已停止
	if ws.Status.Phase == mv1.WorkspacePhaseStopping || ws.Status.Phase == mv1.WorkspacePhaseStopped {
		return res, nil
	}

	// 3.更新Operation字段以停止Workspace
	// 使用Update时,可能由于版本冲突而导致失败,需要重试
	exist := true
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var wp mv1.WorkSpace
		exist = s.checkWorkspaceExist(ctx, client.ObjectKey{Name: name, Namespace: s.namespace}, &wp)
		if !exist {
			return nil
		}

		// 更新workspace的Operation字段
		wp.Spec.Command = mv1.WorkSpaceStop
		if err := s.client.Update(ctx, &wp); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error(err, "update workspace")
		res.Status = pb.ResponseStop_Error
		res.Message = WorkspaceStopFailed
		return res, status.Error(codes.Unknown, err.Error())
	}

	if !exist {
		res.Status = pb.ResponseStop_NotFound
		res.Message = WorkspaceNotExist
		return res, status.Error(codes.NotFound, WorkspaceNotExist)
	}

	return res, nil
}

// RunningWorkspaces 获取运行中的Workspace
func (s *WorkSpaceService) RunningWorkspaces(ctx context.Context, req *pb.RequestRunningWorkspaces) (*pb.ResponseRunningWorkspace, error) {
	res := &pb.ResponseRunningWorkspace{}
	var wss mv1.WorkSpaceList
	err := s.client.List(ctx, &wss, client.MatchingLabels{"uid": req.Uid})
	if err != nil {
		s.logger.Error(err, "list workspace")
		return res, status.Error(codes.Unknown, err.Error())
	}

	// 过滤出正在运行中的Workspace
	for _, item := range wss.Items {
		if item.Status.Phase == mv1.WorkspacePhaseStaring || item.Status.Phase == mv1.WorkspacePhaseRunning {
			res.Workspaces = append(res.Workspaces, &pb.ResponseRunningWorkspace_WorkspaceBasicInfo{
				Sid:  item.Spec.SID,
				Name: "", // 不返回Kubernetes内部名称，让webserver使用数据库中的友好名称
			})
		}
	}

	return res, nil
}

func (s *WorkSpaceService) checkWorkspaceExist(ctx context.Context, key client.ObjectKey, w *mv1.WorkSpace) bool {
	if err := s.client.Get(ctx, key, w); err != nil {
		if errors.IsNotFound(err) {
			return false
		}

		s.logger.Error(err, "get workspace")
		return false
	}

	return true
}

func (s *WorkSpaceService) constructWorkspace(space *pb.RequestCreate, name string) *mv1.WorkSpace {
	hardware := fmt.Sprintf("%sC%s%s", space.ResourceLimit.Cpu,
		strings.Split(space.ResourceLimit.Memory, "i")[0], strings.Split(space.ResourceLimit.Storage, "i")[0])
	return &mv1.WorkSpace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cloud-ide.mangohow.com/v1",
			Kind:       "WorkSpace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.namespace,
			Labels: map[string]string{
				"uid": space.Uid,
				"sid": space.Sid,
			},
		},
		Spec: mv1.WorkSpaceSpec{
			UID:           space.Uid,
			SID:           space.Sid,
			Cpu:           space.ResourceLimit.Cpu,
			Memory:        space.ResourceLimit.Memory,
			Storage:       space.ResourceLimit.Storage,
			Hardware:      hardware,
			Image:         space.Image,
			Port:          space.Port,
			MountPath:     space.VolumeMountPath,
			GitRepository: space.GitRepository,
			Command:       mv1.WorkSpaceStart,
		},
	}
}

func (s *WorkSpaceService) validateRequestCreate(req *pb.RequestCreate) error {
	if len(req.Sid) < 6 || len(req.Sid) > 24 {
		return fmt.Errorf("sid invalid, length of sid must be [6,24], now is%d", len(req.Sid))
	}
	if len(req.Uid) < 6 || len(req.Uid) > 24 {
		return fmt.Errorf("sid invalid, length of uid must be [6,24], now is%d", len(req.Uid))
	}
	if req.Port < 1024 || req.Port > 65535 {
		return fmt.Errorf("port invalid, port must be [1024,65535], now is%d", req.Port)
	}
	if req.GitRepository != "" {
		// 检查是否是环境变量编码格式（Claude模板使用）
		if strings.HasPrefix(req.GitRepository, "ENV:") {
			// 解析编码格式: ENV:{"key":"value"}|GIT:actual_repo_url
			parts := strings.SplitN(req.GitRepository, "|GIT:", 2)
			if len(parts) == 2 {
				// 验证实际的git仓库URL（如果存在）
				actualGitRepo := parts[1]
				if actualGitRepo != "" {
					matched, err := regexp.MatchString(`^https://\S+.git$`, actualGitRepo)
					if err != nil {
						s.logger.Error(err, "regexp")
						return err
					}
					if !matched {
						return fmt.Errorf("git repository invalid")
					}
				}
				// 环境变量编码格式验证通过
			} else {
				return fmt.Errorf("git repository invalid: malformed environment encoding")
			}
		} else {
			// 常规git URL验证
			matched, err := regexp.MatchString(`^https://\S+.git$`, req.GitRepository)
			if err != nil {
				s.logger.Error(err, "regexp")
				return err
			}
			if !matched {
				return fmt.Errorf("git repository invalid")
			}
		}
	}
	matched, err := regexp.MatchString(`^\/(?:[\w-]+\/)*(?:[\w-]+\.[\w-]+|[\w-]+\/?)$`, req.VolumeMountPath)
	if err != nil {
		s.logger.Error(err, "regexp")
		return err
	}
	if !matched {
		return fmt.Errorf("mount path invalid")
	}

	return s.validateResourceLimit(req.ResourceLimit)
}

func (s *WorkSpaceService) validateResourceLimit(limit *pb.ResourceLimit) error {
	_, err := resource.ParseQuantity(limit.Cpu)
	if err != nil {
		return fmt.Errorf("resource limit cpu invalid %s", err.Error())
	}

	_, err = resource.ParseQuantity(limit.Memory)
	if err != nil {
		return fmt.Errorf("resource limit memory invalid %s", err.Error())
	}

	_, err = resource.ParseQuantity(limit.Storage)
	if err != nil {
		return fmt.Errorf("resource limit storage invalid %s", err.Error())
	}

	return nil
}

func workspaceName(uid, sid string) string {
	return fmt.Sprintf(WorkspaceNameFormat, uid, sid)
}
