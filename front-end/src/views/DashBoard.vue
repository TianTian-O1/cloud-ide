
<template>
    <el-container class="root">
      <el-header class="header">
        <div class="header-content">
        <h3 class="site-name">Cloud Code</h3>
          <div class="mobile-menu-btn" @click="toggleMobileMenu" v-if="isMobile">
            <i class="el-icon-menu"></i>
          </div>
          <div class="user-area" :class="{ 'mobile-hidden': isMobile && !showMobileMenu }">
          <div class="avatar-container">
            <img class="user-avatar" :src="user.avatar" alt="">
            <span class="vip-badge" v-if="vipInfo && vipInfo.is_active">VIP</span>
          </div>
          <h4 class="user-nickname">{{user.nickname}}</h4>
          <div class="vip-action-btn" :class="{ 'disabled': !vipInfo || !vipInfo.is_active }" @click="goToVip">
            <span v-if="vipInfo && vipInfo.is_active">VIP管理</span>
            <span v-else>升级VIP</span>
            </div>
          </div>
        </div>
      </el-header>

      <el-container class="bottom-main">
        <el-aside class="aside-menu" :class="{ 'mobile-hidden': isMobile && !showMobileMenu }">
          <el-menu
            :default-active="activePath"
            class="el-menu-vertical-demo"
            :router="true"
            background-color="#303336"  
            text-color="#fff"
            active-text-color="#ffd04b"
            @select="handleMenuSelect">
            <el-menu-item index="/dash/templates">
              <i class="el-icon-menu"></i>
              <span slot="title">空间模板</span>
            </el-menu-item>
            <el-menu-item index="/dash/workspaces">
              <i class="el-icon-document"></i>
              <span slot="title">工作空间</span>
            </el-menu-item>
          </el-menu>
        </el-aside>

        <el-main class="main-content">
          <router-view></router-view>
        </el-main>
      </el-container>

      <!-- 移动端遮罩层 -->
      <div 
        v-if="isMobile && showMobileMenu" 
        class="mobile-overlay" 
        @click="closeMobileMenu">
      </div>
    </el-container>

</template>

<script>

import {Base64} from "js-base64"

export default {
  name: 'DashBoard',
  data() {
    return {
      user: {},
      activePath: "",
      vipInfo: null,
      isMobile: false,
      showMobileMenu: false
    }
  },
  mounted() {
    this.activePath = this.$route.path

    const data = window.sessionStorage.getItem("userData")
    const jdata = Base64.decode(data)
    this.user = JSON.parse(jdata)
    
    // 加载VIP信息
    this.loadVipInfo()
    
    // 检测移动端
    this.checkMobile()
    window.addEventListener('resize', this.checkMobile)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobile)
  },
  methods: {
    async loadVipInfo() {
      try {
        const {data: res} = await this.$axios.get('/api/payment/subscription')
        if (res.status === 0) {
          this.vipInfo = res.data
        }
      } catch (error) {
        console.error('加载VIP信息失败:', error)
      }
    },
    goToVip() {
      // 试用阶段：禁用VIP功能
      if (!this.vipInfo || !this.vipInfo.is_active) {
        this.$message.info('试用阶段暂未开放VIP功能')
        return
      }
      
      this.$router.push('/dash/payment')
      // 移动端点击后关闭菜单
      if (this.isMobile) {
        this.showMobileMenu = false
      }
    },
    checkMobile() {
      this.isMobile = window.innerWidth <= 768
      if (!this.isMobile) {
        this.showMobileMenu = false
      }
    },
    toggleMobileMenu() {
      this.showMobileMenu = !this.showMobileMenu
    },
    closeMobileMenu() {
      this.showMobileMenu = false
    },
    handleMenuSelect() {
      // 移动端选择菜单后自动关闭
      if (this.isMobile) {
        this.showMobileMenu = false
      }
    }
  }
}
</script>

<style lang="less" scoped>

.root {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.el-header {
  background-color: #373b42;
  color: #333;
  line-height: 60px;
  padding: 0;
  flex-shrink: 0;

  .header-content {
    height: 100%;
    padding: 0 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    position: relative;

    @media (min-width: 769px) {
  padding: 0 80px;
    }
  }

  .site-name {
    color: rgb(75, 196, 165);
    margin: 0;
    font-size: 20px;
    font-weight: 550;
    font-family: Georgia;

    @media (min-width: 769px) {
      font-size: 24px;
    }
  }

  .mobile-menu-btn {
    display: none;
    color: #fff;
    font-size: 20px;
    cursor: pointer;
    padding: 10px;
    border-radius: 4px;
    transition: background-color 0.3s;

    &:hover {
      background-color: rgba(255, 255, 255, 0.1);
    }

    @media (max-width: 768px) {
      display: block;
    }
  }

  .user-area {
    display: flex;
    align-items: center;
    cursor: pointer;
    gap: 8px;

    @media (max-width: 768px) {
      position: absolute;
      top: 100%;
      right: 20px;
      background: #373b42;
      border: 1px solid #555;
      border-radius: 8px;
      padding: 12px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.3);
      z-index: 1001;
      gap: 6px;
      min-width: 160px;
      flex-direction: row;
      flex-wrap: wrap;
      justify-content: center;
      
      &.mobile-hidden {
        display: none;
      }
    }
  }

  .avatar-container {
    position: relative;
    flex-shrink: 0;
  }

  .user-avatar {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: 2px solid #fff;

    @media (max-width: 768px) {
      width: 36px;
      height: 36px;
    }
  }

  .vip-badge {
    position: absolute;
    top: -3px;
    right: -3px;
    background: linear-gradient(45deg, #FFD700, #FFA500);
    color: #fff;
    font-size: 6px;
    font-weight: bold;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid #fff;
    box-shadow: 0 1px 2px rgba(0,0,0,0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10;

    @media (max-width: 768px) {
      font-size: 8px;
      width: 14px;
      height: 14px;
    }
  }

  .user-nickname {
    color: #fff;
    margin: 0;
    font-weight: normal;
    font-size: 16px;
    white-space: nowrap;

    @media (max-width: 768px) {
      font-size: 13px;
      width: 100%;
      text-align: center;
      margin: 2px 0;
    }
  }

  .vip-action-btn {
    padding: 6px 12px;
    background: linear-gradient(45deg, #4bc4a5, #36b491);
    color: #fff;
    border-radius: 15px;
    font-size: 12px;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.3s ease;
    white-space: nowrap;
    flex-shrink: 0;

    &:hover:not(.disabled) {
      background: linear-gradient(45deg, #36b491, #2da17d);
      transform: translateY(-1px);
      box-shadow: 0 2px 6px rgba(0,0,0,0.3);
    }

    &.disabled {
      background: #999;
      color: #ccc;
      cursor: not-allowed;
      opacity: 0.6;
      
      &:hover {
        background: #999;
        transform: none;
        box-shadow: none;
      }
    }

    @media (max-width: 768px) {
      padding: 6px 10px;
      font-size: 11px;
      margin-top: 6px;
      width: auto;
      min-width: 70px;
      text-align: center;
      border-radius: 12px;
    }
  }
}

.bottom-main {
  flex: 1;
  display: flex;
  overflow: visible; /* 允许内容自然显示 */
  min-height: 0; /* 重置最小高度约束 */
}

.aside-menu {
  background-color: #303336;
  width: 200px !important;
  flex-shrink: 0;
  transition: all 0.3s ease;

  .el-menu {
    border: 0 !important;
    height: 100%;
}

  @media (max-width: 768px) {
    position: fixed;
    top: 60px;
    left: 0;
    height: auto;
    max-height: calc(100vh - 80px);
    width: 180px !important;
    z-index: 1000;
    transform: translateX(0);
    border-radius: 0 12px 12px 0;
    box-shadow: 4px 0 15px rgba(0,0,0,0.3);
    
    &.mobile-hidden {
      transform: translateX(-100%);
    }
    
    .el-menu {
      padding: 15px 0 10px 0;
      border-radius: 0 12px 12px 0;
      min-height: auto;
      height: auto;
    }
  }

  @media (min-width: 769px) {
    width: 150px !important;
  }
}

.main-content {
  flex: 1;
  padding: 0;
  background-color: rgb(33, 35, 41);
  overflow: visible; /* 允许内容自然显示 */
  min-height: 0; /* 重置最小高度 */

  @media (max-width: 768px) {
    width: 100%;
  }
  }

.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.3);
  z-index: 999;
  transition: opacity 0.3s ease;

  @media (min-width: 769px) {
    display: none;
  }
}

/* 优化触摸目标大小 */
@media (max-width: 768px) {
  .el-menu-item {
    height: 48px !important;
    line-height: 48px !important;
    font-size: 14px !important;
    margin: 0 8px 6px 8px !important;
    border-radius: 8px !important;
    transition: all 0.3s ease !important;
    
    &:hover {
      background-color: rgba(255, 255, 255, 0.1) !important;
}

    &.is-active {
      background-color: #4bc4a5 !important;
      color: #fff !important;
      
      i {
        color: #fff !important;
      }
    }
    
    i {
      font-size: 16px !important;
      margin-right: 8px !important;
    color: #fff;
    }
    
    span {
      color: #fff !important;
      font-weight: 500;
    }
  }
  
  /* 移动端菜单容器优化 */
  .aside-menu {
    .el-menu {
      background-color: transparent !important;
  }
}
}

/* 解决Element UI在移动端的一些问题 */
@media (max-width: 768px) {
  :deep(.el-menu-item) {
    padding-left: 20px !important;
    padding-right: 20px !important;
  }
  
  :deep(.el-menu-item span) {
    font-size: 16px !important;
  }
}

</style>