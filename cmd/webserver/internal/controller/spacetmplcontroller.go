package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/mangohow/cloud-ide/pkg/utils"
	"github.com/sirupsen/logrus"
)

type SpaceTmplController struct {
	logger              *logrus.Logger
	service             *service.SpaceTmplService
	subscriptionService *service.SubscriptionService
}

func NewSpaceTmplController() *SpaceTmplController {
	return &SpaceTmplController{
		logger:              logger.Logger(),
		service:             service.NewSpaceTmplService(),
		subscriptionService: service.NewSubscriptionService(),
	}
}

// SpaceTmpls 获取所有模板 method: GET path:/api/template/list
// 根据用户VIP状态过滤模板列表
func (s *SpaceTmplController) SpaceTmpls(ctx *gin.Context) *serialize.Response {
	tmpls, kinds, err := s.service.GetAllUsingTmpl()
	if err != nil {
		s.logger.Warnf("get tmpls err:%v", err)
		return serialize.Fail(code.QueryFailed)
	}

	// 获取用户ID并检查VIP状态
	userId := utils.MustGet[uint32](ctx, "id")
	isVip := s.subscriptionService.IsUserVip(userId)
	
	// 试用阶段：普通用户也可以访问Claude模板，移除过滤逻辑
	// if !isVip {
	//	 filteredTmpls := make([]*model.SpaceTemplate, 0)
	//	 for _, tmpl := range tmpls {
	//		 if tmpl.Id != 7 { // 排除Claude模板
	//			 filteredTmpls = append(filteredTmpls, tmpl)
	//		 }
	//	 }
	//	 tmpls = filteredTmpls
	//	 s.logger.Infof("普通用户 %d 获取模板列表，已过滤Claude模板", userId)
	// }
	
	if !isVip {
		s.logger.Infof("普通用户 %d 获取模板列表，试用阶段开放所有模板", userId)
	}

	return serialize.OkData(gin.H{
		"tmpls": tmpls,
		"kinds": kinds,
	})
}

// SpaceSpecs 获取空间规格 method: GET path:/api/spec/list
// 根据用户VIP状态过滤规格列表
func (s *SpaceTmplController) SpaceSpecs(ctx *gin.Context) *serialize.Response {
	specs, err := s.service.GetAllSpec()
	if err != nil {
		s.logger.Warnf("get specs error:%v", err)
		return serialize.Fail(code.QueryFailed)
	}

	// 获取用户ID并检查VIP状态
	userId := utils.MustGet[uint32](ctx, "id")
	isVip := s.subscriptionService.IsUserVip(userId)
	
	// 如果不是VIP用户，只返回测试型规格（ID=4）
	if !isVip {
		filteredSpecs := make([]*model.SpaceSpec, 0)
		for _, spec := range specs {
			if spec.Id == 4 { // 只保留测试型规格
				filteredSpecs = append(filteredSpecs, spec)
			}
		}
		specs = filteredSpecs
		s.logger.Infof("普通用户 %d 获取规格列表，只返回测试型规格", userId)
	}

	return serialize.OkData(specs)
}
