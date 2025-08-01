package controller

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model/reqtype"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/mangohow/cloud-ide/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	// 普通用户限制常量
	TestSpecId   = uint32(4) // 测试型规格ID
	ClaudeTmplId = uint32(7) // Claude模板ID
)

var (
	// 权限错误
	ErrPermissionDeniedSpec = errors.New("普通用户只能创建测试型配置的工作空间，请升级为VIP用户使用其他配置")
	ErrPermissionDeniedTemplate = errors.New("Claude AI助手功能仅限VIP用户使用，请升级为VIP用户")
)

type CloudCodeController struct {
	logger              *logrus.Logger
	spaceService        *service.CloudCodeService
	subscriptionService *service.SubscriptionService
}

func NewCloudCodeController() *CloudCodeController {
	return &CloudCodeController{
		logger:              logger.Logger(),
		spaceService:        service.NewCloudCodeService(),
		subscriptionService: service.NewSubscriptionService(),
	}
}

// CreateSpace 创建一个云空间  method: POST path: /api/workspace
// Request Param: reqtype.SpaceCreateOption
func (c *CloudCodeController) CreateSpace(ctx *gin.Context) *serialize.Response {
	// 1、用户参数获取和验证
	req, err := c.creationCheck(ctx)
	if err != nil {
		if err == ErrPermissionDeniedSpec || err == ErrPermissionDeniedTemplate {
			return serialize.NewResponse(http.StatusForbidden, code.QueryFailed, nil, err.Error())
		}
		return serialize.Error(http.StatusBadRequest)
	}
	if req == nil {
		return serialize.Error(http.StatusBadRequest)
	}

	// 2、获取用户id，在token验证时已经解析出并放入ctx中了
	userId := utils.MustGet[uint32](ctx, "id")

	// 3、调用service处理然后响应结果
	space, err := c.spaceService.CreateWorkspace(req, userId)
	switch err {
	case service.ErrNameDuplicate:
		return serialize.Fail(code.SpaceCreateNameDuplicate)
	case service.ErrReachMaxSpaceCount:
		return serialize.Fail(code.SpaceCreateReachMaxCount)
	case service.ErrSpaceCreate:
		return serialize.Fail(code.SpaceCreateFailed)
	case service.ErrReqParamInvalid:
		return serialize.Error(http.StatusBadRequest)
	}

	if err != nil {
		return serialize.Fail(code.SpaceCreateFailed)
	}

	return serialize.OkData(space)
}

// creationCheck 用户参数验证
func (c *CloudCodeController) creationCheck(ctx *gin.Context) (*reqtype.SpaceCreateOption, error) {
	// 获取用户请求参数
	var req reqtype.SpaceCreateOption
	// 绑定数据
	err := ctx.ShouldBind(&req)
	if err != nil {
		return nil, err
	}

	c.logger.Debug(req)

	if req.GitRepository != "" {
		matched, err := regexp.MatchString(`^https://\S+.git$`, req.GitRepository)
		if err != nil {
			c.logger.Error("regexp invalid")
			return nil, err
		}
		if !matched {
			c.logger.Error("git repository invalid")
			return nil, errors.New("git repository invalid")
		}
	}

	// 参数验证
	get1, exist1 := ctx.Get("id")
	_, exist2 := ctx.Get("username")
	if !exist1 || !exist2 {
		return nil, errors.New("user not found")
	}
	id, ok := get1.(uint32)
	if !ok || id != req.UserId {
		return nil, errors.New("user id mismatch")
	}

	// 检查用户权限
	isVip := c.subscriptionService.IsUserVip(req.UserId)
	if !isVip {
		// 普通用户权限限制（试用阶段开放Claude模板）
		if req.SpaceSpecId != TestSpecId {
			c.logger.Warnf("普通用户尝试使用非测试型规格: spec_id=%d, user_id=%d", req.SpaceSpecId, req.UserId)
			return nil, ErrPermissionDeniedSpec
		}
		
		// 试用阶段：普通用户也可以使用Claude模板
		// if req.TmplId == ClaudeTmplId {
		//	c.logger.Warnf("普通用户尝试使用Claude模板: tmpl_id=%d, user_id=%d", req.TmplId, req.UserId)
		//	return nil, ErrPermissionDeniedTemplate
		// }
	}

	return &req, nil
}

// CreateSpaceAndStart 创建一个新的云空间并启动 method: POST path: /api/space_cas
// Request Param: reqtype.SpaceCreateOption
func (c *CloudCodeController) CreateSpaceAndStart(ctx *gin.Context) *serialize.Response {
	req, err := c.creationCheck(ctx)
	if err != nil {
		if err == ErrPermissionDeniedSpec || err == ErrPermissionDeniedTemplate {
			return serialize.NewResponse(http.StatusForbidden, code.QueryFailed, nil, err.Error())
		}
		return serialize.Error(http.StatusBadRequest)
	}
	if req == nil {
		return serialize.Error(http.StatusBadRequest)
	}

	userId := utils.MustGet[uint32](ctx, "id")
	uid := utils.MustGet[string](ctx, "uid")

	space, err := c.spaceService.CreateAndStartWorkspace(req, userId, uid)
	switch err {
	case service.ErrNameDuplicate:
		return serialize.Fail(code.SpaceCreateNameDuplicate)
	case service.ErrReachMaxSpaceCount:
		return serialize.Fail(code.SpaceCreateReachMaxCount)
	case service.ErrSpaceCreate:
		return serialize.Fail(code.SpaceCreateFailed)
	case service.ErrSpaceStart:
		return serialize.Fail(code.SpaceStartFailed)
	case service.ErrOtherSpaceIsRunning:
		return serialize.Fail(code.SpaceOtherSpaceIsRunning)
	case service.ErrReqParamInvalid:
		return serialize.Error(http.StatusBadRequest)
	case service.ErrSpaceAlreadyExist:
		return serialize.Fail(code.SpaceAlreadyExist)
	case service.ErrResourceExhausted:
		return serialize.Fail(code.ResourceExhausted)
	}

	if err != nil {
		return serialize.Fail(code.SpaceCreateFailed)
	}

	return serialize.OkData(space)
}

// StartSpace 启动一个已存在的云空间 method: POST path: /api/workspace/start
// request param: space id
func (c *CloudCodeController) StartSpace(ctx *gin.Context) *serialize.Response {
	var req reqtype.SpaceId
	err := ctx.ShouldBind(&req)
	if err != nil {
		c.logger.Warnf("bind param error:%v", err)
		return serialize.Error(http.StatusBadRequest)
	}

	userId := utils.MustGet[uint32](ctx, "id")
	uid := utils.MustGet[string](ctx, "uid")

	space, err := c.spaceService.StartWorkspace(req.Id, userId, uid)
	switch err {
	case service.ErrWorkSpaceNotExist:
		return serialize.Fail(code.SpaceStartNotExist)
	case service.ErrSpaceStart:
		return serialize.Fail(code.SpaceStartFailed)
	case service.ErrOtherSpaceIsRunning:
		return serialize.Fail(code.SpaceOtherSpaceIsRunning)
	case service.ErrSpaceNotFound:
		return serialize.Fail(code.SpaceNotFound)
	}

	if err != nil {
		return serialize.Fail(code.SpaceStartFailed)
	}

	return serialize.OkData(space)
}

// StopSpace 停止正在运行的云空间 method: PUT path: /api/workspace/stop
// Request Param: sid
func (c *CloudCodeController) StopSpace(ctx *gin.Context) *serialize.Response {
	var req reqtype.SpaceId
	err := ctx.ShouldBind(&req)
	if err != nil {
		c.logger.Warnf("bind param error:%v", err)
		return serialize.Error(http.StatusBadRequest)
	}

	uid := utils.MustGet[string](ctx, "uid")
	userId := utils.MustGet[uint32](ctx, "id")

	err = c.spaceService.StopWorkspace(req.Id, userId, uid)
	if err != nil {
		if err == service.ErrWorkSpaceIsNotRunning {
			return serialize.Ok()
		}

		return serialize.Fail(code.SpaceStopFailed)
	}

	return serialize.Ok()
}

// DeleteSpace 删除已存在的云空间  method: DELETE path: /api/workspace
// Request Param: id
func (c *CloudCodeController) DeleteSpace(ctx *gin.Context) *serialize.Response {
	var req reqtype.SpaceId
	err := ctx.ShouldBind(&req)
	if err != nil {
		c.logger.Warnf("bind param error:%v", err)
		return serialize.Error(http.StatusBadRequest)
	}
	c.logger.Debug("space id:", req.Id)

	// 获取用户id和用户uid
	userId := utils.MustGet[uint32](ctx, "id")
	uid := utils.MustGet[string](ctx, "uid")

	err = c.spaceService.DeleteWorkspace(req.Id, userId, uid)
	if err != nil {
		if err == service.ErrWorkSpaceIsRunning {
			return serialize.Fail(code.SpaceDeleteIsRunning)
		}

		return serialize.Fail(code.SpaceDeleteFailed)
	}

	return serialize.Ok()
}

// ListSpace 获取所有创建的云空间 method: GET path: /api/workspace/list
// Request param: id uid
func (c *CloudCodeController) ListSpace(ctx *gin.Context) *serialize.Response {
	userId := utils.MustGet[uint32](ctx, "id")
	uid := utils.MustGet[string](ctx, "uid")

	spaces, err := c.spaceService.ListWorkspace(userId, uid)
	if err != nil {
		return serialize.Fail(code.QueryFailed)
	}

	return serialize.OkData(spaces)
}

// ModifySpaceName 修改工作空间名称 method: POST path: /api/workspace/name
func (c *CloudCodeController) ModifySpaceName(ctx *gin.Context) *serialize.Response {
	var req struct {
		Name string `json:"name"` // 新的工作空间的名称
		Id   uint32 `json:"id"`   // 工作空间id
	}
	err := ctx.ShouldBind(&req)
	if err != nil {
		c.logger.Warnf("bind req error:%v", err)
		return serialize.Fail(code.SpaceNameModifyFailed)
	}

	userId := utils.MustGet[uint32](ctx, "id")

	err = c.spaceService.ModifyName(req.Name, req.Id, userId)
	switch err {
	case service.ErrNameDuplicate:
		return serialize.Fail(code.SpaceCreateNameDuplicate)
	case nil:
		return serialize.Ok()
	default:
		return serialize.Fail(code.SpaceNameModifyFailed)
	}
}
