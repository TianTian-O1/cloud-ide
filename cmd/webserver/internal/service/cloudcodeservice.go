package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/caches"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model/reqtype"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/rpc"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
	"gopkg.in/mgo.v2/bson"
)

const (
	DefaultPodPort = 9999
	MaxSpaceCount  = 10
)

type CloudCodeService struct {
	logger    *logrus.Logger
	rpc       pb.CloudIdeServiceClient
	dao       *dao.SpaceDao
	tmplCache *caches.TmplCache
	specCache *caches.SpecCache
}

func NewCloudCodeService() *CloudCodeService {
	conn := rpc.GrpcClient("space-code")
	factory := caches.CacheFactory()
	d := dao.NewSpaceTemplateDao()
	return &CloudCodeService{
		logger:    logger.Logger(),
		rpc:       pb.NewCloudIdeServiceClient(conn),
		dao:       dao.NewSpaceDao(),
		tmplCache: factory.TmplCache(d),
		specCache: factory.SpecCache(d),
	}
}

var (
	ErrReqParamInvalid    = errors.New("request param invalid")
	ErrNameDuplicate      = errors.New("name duplicate")
	ErrReachMaxSpaceCount = errors.New("reach max space count")
	ErrSpaceCreate        = errors.New("space create failed")
	ErrSpaceStart         = errors.New("space start failed")
	ErrSpaceDelete        = errors.New("space delete failed")
	ErrSpaceStop          = errors.New("space stop failed")
	ErrSpaceAlreadyExist  = errors.New("space already exist")
	ErrSpaceNotFound      = errors.New("space not found")
	ErrResourceExhausted  = errors.New("no adequate resource are available")
)

// CreateWorkspace 创建云工作空间, 只在数据库中插入一条记录
func (c *CloudCodeService) CreateWorkspace(req *reqtype.SpaceCreateOption, userId uint32) (*model.Space, error) {
	// 1、验证创建的工作空间是否达到最大数量
	count, err := c.dao.FindCountByUserId(userId)
	if err != nil {
		c.logger.Warnf("get space count error:%v", err)
		return nil, ErrSpaceCreate
	}
	if count >= MaxSpaceCount {
		return nil, ErrReachMaxSpaceCount
	}

	// 2、验证名称是否重复
	if err := c.dao.FindByUserIdAndName(userId, req.Name); err == nil {
		c.logger.Warnf("find space error:%v", err)
		return nil, ErrNameDuplicate
	}

	// 3、从缓存中获取要创建的云空间的模板
	tmpl := c.tmplCache.GetTmpl(req.TmplId)
	if tmpl == nil {
		c.logger.Warnf("get tmpl cache error:%v", err)
		return nil, ErrReqParamInvalid
	}

	// 4、从缓存中获取要创建的云空间的规格
	spec := c.specCache.Get(req.SpaceSpecId)
	if spec == nil {
		return nil, ErrReqParamInvalid
	}

	// 5、构造云工作空间结构
	now := time.Now()
	
	// 构建环境变量配置（特别是Claude模板）
	envConfig := ""
	if tmpl.Id == 7 {
		// Claude模板支持多AI提供商配置
		envVars := make(map[string]string)
		
		// Anthropic API 配置
		if req.AnthropicAuthToken != "" {
			envVars["ANTHROPIC_AUTH_TOKEN"] = req.AnthropicAuthToken
		}
		if req.AnthropicBaseURL != "" {
			envVars["ANTHROPIC_BASE_URL"] = req.AnthropicBaseURL
		}
		
		// OpenAI API 配置
		if req.OpenAIAPIKey != "" {
			envVars["OPENAI_API_KEY"] = req.OpenAIAPIKey
		}
		if req.OpenAIBaseURL != "" {
			envVars["OPENAI_BASE_URL"] = req.OpenAIBaseURL
		}
		
		// DeepSeek API 配置
		if req.DeepSeekAPIKey != "" {
			envVars["DEEPSEEK_API_KEY"] = req.DeepSeekAPIKey
		}
		
		// Gemini API 配置
		if req.GeminiAPIKey != "" {
			envVars["GEMINI_API_KEY"] = req.GeminiAPIKey
		}
		
		// Moonshot API 配置
		if req.MoonshotAPIKey != "" {
			envVars["MOONSHOT_API_KEY"] = req.MoonshotAPIKey
		}
		
		// Qwen API 配置
		if req.QwenAPIKey != "" {
			envVars["QWEN_API_KEY"] = req.QwenAPIKey
		}
		
		// 模型配置
		if req.BigModel != "" {
			envVars["BIG_MODEL"] = req.BigModel
		} else {
			envVars["BIG_MODEL"] = "claude-3-5-sonnet-20241022" // 默认大模型
		}
		if req.SmallModel != "" {
			envVars["SMALL_MODEL"] = req.SmallModel
		} else {
			envVars["SMALL_MODEL"] = "claude-3-haiku-20240307" // 默认小模型
		}
		
		if envData, err := json.Marshal(envVars); err == nil {
			envConfig = string(envData)
		}
	}
	
	space := &model.Space{
		UserId:        userId,
		TmplId:        tmpl.Id,
		SpecId:        spec.Id,
		Spec:          *spec,
		Name:          req.Name,
		Status:        model.SpaceStatusUncreated,
		CreateTime:    now,
		DeleteTime:    now,
		StopTime:      now,
		TotalTime:     0,
		Sid:           generateSID(),
		GitRepository: req.GitRepository,
		Environment:   envConfig,
	}

	// 6、 添加到数据库
	spaceId, err := c.dao.Insert(space)
	if err != nil {
		c.logger.Errorf("add space error:%v", err)
		return nil, ErrSpaceCreate
	}
	space.Id = spaceId

	return space, nil
}

var ErrOtherSpaceIsRunning = errors.New("there is other space running")

func (c *CloudCodeService) checkHasRunningWorkspace(uid string) (bool, error) {
	// 1、检查是否有其它工作空间正在运行, 同时只能有一个工作空间启动
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	wss, err := c.rpc.RunningWorkspaces(ctx, &pb.RequestRunningWorkspaces{Uid: uid})
	if err != nil {
		c.logger.Errorf("get running workspaces err=%v", err)
		return true, err
	}
	c.logger.Debug("running workspaces:", wss.Workspaces)
	if len(wss.Workspaces) > 0 {
		return true, nil
	}

	return false, nil
}

// CreateAndStartWorkspace 创建并且启动云工作空间
func (c *CloudCodeService) CreateAndStartWorkspace(req *reqtype.SpaceCreateOption, userId uint32, uid string) (*model.Space, error) {
	// 1、检查是否有其它工作空间正在运行, 同时只能有一个工作空间启动
	// 临时禁用此检查以修复竞态条件问题
	// TODO: 需要改进逻辑，避免在创建过程中检查刚创建的工作空间
	/*
	if ok, err := c.checkHasRunningWorkspace(uid); err != nil || ok {
		if err != nil {
			return nil, ErrSpaceCreate
		}
		return nil, ErrOtherSpaceIsRunning
	}
	*/

	// 2、创建工作空间
	space, err := c.CreateWorkspace(req, userId)
	if err != nil {
		return nil, err
	}

	// 3、真正的创建并且启动工作空间
	return c.createAndStartWorkspace(space, uid)
}

// 调用rpc来创建并且启动工作空间
func (c *CloudCodeService) createAndStartWorkspace(space *model.Space, uid string) (*model.Space, error) {
	// 1、获取空间模板
	tmpl := c.tmplCache.GetTmpl(space.TmplId)
	if tmpl == nil {
		c.logger.Warnf("get tmpl cache error")
		return nil, ErrSpaceStart
	}

	// 2、获取硬件规格
	spec := c.specCache.Get(space.SpecId)
	if spec == nil {
		c.logger.Errorf("get spec cache error")
		return nil, ErrSpaceStart
	}

	// 3、生成Workspace信息
	gitRepo := space.GitRepository
	
	// 如果是Claude模板且有环境变量配置，将环境变量编码到GitRepository字段中
	if tmpl.Id == 7 && space.Environment != "" {
		// 使用特殊前缀标识包含环境变量的GitRepository
		gitRepo = "ENV:" + space.Environment + "|GIT:" + space.GitRepository
	}
	
	ws := &pb.RequestCreate{
		Sid:             space.Sid,
		Uid:             uid,
		Image:           tmpl.Image,
		Port:            DefaultPodPort,
		GitRepository:   gitRepo,
		VolumeMountPath: "/root/",
		ResourceLimit: &pb.ResourceLimit{
			Cpu:     spec.CpuSpec,
			Memory:  spec.MemSpec,
			Storage: spec.StorageSpec,
		},
	}

	c.logger.Debug(ws.ResourceLimit)

	var retErr error
	// 4、请求k8s controller创建并启动云空间
	// 设置90分钟的超时时间
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*90)
	defer cancelFunc()
	resp, err := c.rpc.CreateSpace(ctx, ws)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return nil, err
		}
		c.logger.Error("create workspace err=", s.Message())

		c.logger.Debug("resp:", resp)

		// 只有在resp不为nil时才检查状态
		if resp != nil {
			switch resp.Status {
			case pb.ResponseCreate_AlreadyExist:
				return nil, ErrSpaceAlreadyExist
			case pb.ResponseCreate_Error:
				return nil, ErrSpaceCreate
			}
		}
		// 如果resp为nil或者其他错误状态，返回通用错误
		return nil, err
	}

	space.RunningStatus = model.RunningStatusRunning
	// 5、修改数据库中的状态信息
	if space.Status == model.SpaceStatusUncreated {
		// 更新数据库
		err := c.dao.UpdateStatusById(space.Id, model.SpaceStatusAvailable)
		if err != nil {
			c.logger.Warnf("update space status error:%v", err)
		}
	}

	if retErr != nil {
		return nil, retErr
	}

	return space, nil
}

var ErrWorkSpaceNotExist = errors.New("workspace is not exist")

// StartWorkspace 启动云工作空间
func (c *CloudCodeService) StartWorkspace(id, userId uint32, uid string) (*model.Space, error) {
	// 1、检查是否有其它工作空间正在运行, 同时只能有一个工作空间启动
	// 临时禁用此检查以修复竞态条件问题
	/*
	if ok, err := c.checkHasRunningWorkspace(uid); err != nil || ok {
		if err != nil {
			return nil, ErrSpaceStart
		}
		return nil, ErrOtherSpaceIsRunning
	}
	*/

	// 2.查询该工作空间是否存在
	space, err := c.dao.FindByIdAndUserId(id, userId)
	if err != nil {
		c.logger.Warnf("find space error:%v", err)
		return nil, ErrWorkSpaceNotExist
	}
	space.Id = id
	space.UserId = userId

	// 3.该工作空间是否是第一次启动
	switch space.Status {
	case model.SpaceStatusDeleted:
		return nil, ErrWorkSpaceNotExist
	case model.SpaceStatusUncreated:
		// 这种情况是工作空间被创建时，只插入了数据库
		// 并没有在workspace controller 创建
		// 因此需要创建并且启动
		return c.createAndStartWorkspace(space, uid)
	}

	// 4.启动工作空间
	return c.startWorkspace(space, uid)
}

// startWorkspace 启动工作空间
func (c *CloudCodeService) startWorkspace(space *model.Space, uid string) (*model.Space, error) {
	// 1、获取空间模板
	tmpl := c.tmplCache.GetTmpl(space.TmplId)
	if tmpl == nil {
		c.logger.Errorf("get tmpl cache error")
		return nil, ErrSpaceStart
	}

	// 2、获取硬件规格
	spec := c.specCache.Get(space.SpecId)
	if spec == nil {
		c.logger.Errorf("get spec cache error")
		return nil, ErrSpaceStart
	}

	// 3、生成请求信息
	req := &pb.RequestStart{
		Sid: space.Sid,
		Uid: uid,
		ResourceLimit: &pb.ResourceLimit{
			Cpu:     spec.CpuSpec,
			Memory:  spec.MemSpec,
			Storage: spec.StorageSpec,
		},
	}

	// 4、请求k8s controller启动云空间
	// 设置90s的超时时间
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*90)
	defer cancelFunc()
	resp, err := c.rpc.StartSpace(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return nil, err
		}
		c.logger.Errorf("start workspace err=%s sid=%s", s.Message(), req.Sid)

		switch resp.Status {
		// 工作空间不存在
		case pb.ResponseStart_NotFound:
			return nil, ErrSpaceNotFound
		// 启动工作空间时,工作空间不存在
		case pb.ResponseStart_Error:
			return nil, ErrSpaceStart
		}
	}

	return space, nil
}

var (
	ErrWorkSpaceIsRunning    = errors.New("workspace is running")
	ErrWorkSpaceIsNotRunning = errors.New("workspace is not running")
)

// DeleteWorkspace 删除云工作空间
func (c *CloudCodeService) DeleteWorkspace(id, userId uint32, uid string) error {
	// 1、先查询工作空间并确保该工作空间是属于该用户的
	space, err := c.dao.FindByIdAndUserId(id, userId)
	if err != nil {
		c.logger.Warnf("find sid error:%v", err)
		return err
	}

	// 2.检测是否正在运行
	// 临时禁用此检查以修复竞态条件问题
	/*
	if ok, err := c.checkHasRunningWorkspace(uid); err != nil || ok {
		if err != nil {
			return ErrSpaceDelete
		}
		return ErrWorkSpaceIsRunning
	}
	*/

	// 3、通知controller删除该workspace关联的资源
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	_, err = c.rpc.DeleteSpace(ctx, &pb.RequestDelete{
		Sid: space.Sid,
		Uid: uid,
	})
	if err != nil {
		c.logger.Warnf("delete workspace err:%v", err)
		return err
	}

	// 3、从mysql中删除记录
	return c.dao.DeleteSpaceById(id)
}

// StopWorkspace 停止云工作空间
func (c *CloudCodeService) StopWorkspace(id, userId uint32, uid string) error {
	c.logger.Debugf("StopWorkspace, sid: %d, uid: %s", id, uid)

	// 1、检测该工作空间是否属于该用户
	// TODO 可优化的点，使用redis缓存用户工作空间
	space, err := c.dao.FindByIdAndUserId(id, userId)
	if err != nil {
		c.logger.Warnf("find sid error:%v", err)
		return err
	}

	// 2、查询云工作空间是否正在运行
	ok, err := c.checkHasRunningWorkspace(uid)
	if err != nil {
		c.logger.Errorf("get running workspace err=%v, sid=%d", err, id)
		return ErrSpaceStop
	}
	if !ok {
		c.logger.Debug("workspace is not running, sid:", id)
		return ErrWorkSpaceIsNotRunning
	}

	// 3、停止workspace
	_, err = c.rpc.StopSpace(context.Background(), &pb.RequestStop{
		Sid: space.Sid,
		Uid: uid,
	})
	if err != nil {
		c.logger.Errorf("rpc delete space error:%v", err)
		return err
	}

	return nil
}

// ListWorkspace 列出云工作空间
func (c *CloudCodeService) ListWorkspace(userId uint32, uid string) ([]model.Space, error) {
	spaces, err := c.dao.FindAllSpaceByUserId(userId)
	if err != nil {
		c.logger.Warnf("find spaces error:%v", err)
		return nil, err
	}

	// 填充environment字段和spec字段
	for i := 0; i < len(spaces); i++ {
		t := c.tmplCache.GetTmpl(spaces[i].TmplId)
		spaces[i].Environment = t.Desc
		spaces[i].Avatar = t.Avatar
		spaces[i].Spec = *c.specCache.Get(spaces[i].SpecId)
		spaces[i].Spec.Id = 0
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()
	wss, err := c.rpc.RunningWorkspaces(ctx, &pb.RequestRunningWorkspaces{Uid: uid})
	if err != nil {
		c.logger.Warnf("get running space error:%v, uid:%s", err, uid)
		return spaces, nil
	}

	for i, item := range spaces {
		for _, ws := range wss.Workspaces {
			if item.Sid == ws.Sid {
				spaces[i].RunningStatus = model.RunningStatusRunning
				break
			}
		}

	}

	return spaces, nil
}

func (c *CloudCodeService) ModifyName(name string, id, userId uint32) error {
	// 1、验证名称是否重复
	if err := c.dao.FindByUserIdAndName(userId, name); err == nil {
		c.logger.Warnf("find space error:%v", err)
		return ErrNameDuplicate
	}

	// 2.修改名称
	err := c.dao.UpdateNameById(name, id)
	if err != nil {
		c.logger.Warnf("update space name error:%v", err)
		return err
	}

	return nil
}

// generateSID 生成Space id
func generateSID() string {
	return bson.NewObjectId().Hex()
}
