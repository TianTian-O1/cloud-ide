<template>
  <div class="template-card" :class="{ 'vip-template': isVipTemplate, 'disabled': isDisabled }" @click="handleClick">
    <!-- VIP Badge -->
    <div v-if="isVipTemplate" class="vip-badge">
      <i class="el-icon-crown"></i>
      VIP
    </div>
    
    <!-- Template Image/Icon -->
    <div class="card-header">
      <div class="template-icon">
        <img v-if="info.avatar" :src="info.avatar" :alt="info.name" class="icon-image">
        <div v-else class="icon-placeholder">
          <i class="el-icon-document"></i>
        </div>
      </div>
    </div>
    
    <!-- Template Information -->
    <div class="card-body">
      <h3 class="template-name">{{ info.name }}</h3>
      <p class="template-desc">{{ info.desc }}</p>
      
      <!-- Tags -->
      <div class="template-tags">
        <span v-for="(tag, index) in info.tags" :key="index" class="tag">
          {{ tag }}
        </span>
      </div>
    </div>
    
    <!-- Action Area -->
    <div class="card-footer">
      <div class="action-button" :class="{ 'disabled': isDisabled }">
        <i v-if="isDisabled" class="el-icon-lock"></i>
        <span v-if="isDisabled">需要VIP</span>
        <span v-else>创建空间</span>
      </div>
    </div>
    
    <!-- Hover Effect Overlay -->
    <div class="hover-overlay"></div>
  </div>
</template>

<script>
export default {
  name: 'TemplateCard',
  props: {
    info: {
      type: Object,
      required: true
    },
    vipInfo: {
      type: Object,
      default: () => ({ is_active: false })
    }
  },
  computed: {
    isVipTemplate() {
      // 试用阶段：Claude模板开放给所有用户
      return false; // this.info.id === 7;
    },
    isDisabled() {
      return this.isVipTemplate && !this.vipInfo.is_active;
    }
  },
  methods: {
    handleClick() {
      if (!this.isDisabled) {
        this.$emit('select', this.info.id);
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.template-card {
  position: relative;
  background: linear-gradient(145deg, #3a3f47 0%, #2d3238 100%);
  border-radius: 16px;
  padding: 24px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #404854;
  overflow: hidden;
  height: 280px;
  display: flex;
  flex-direction: column;
  
  &:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
    border-color: #5a6c7d;
    
    .hover-overlay {
      opacity: 1;
    }
    
    .action-button:not(.disabled) {
      background: linear-gradient(135deg, #409eff 0%, #67c23a 100%);
      color: white;
    }
  }
  
  &.vip-template {
    border-color: #ffd700;
    background: linear-gradient(145deg, #4a4237 0%, #3d3b2d 100%);
    
    &:hover {
      border-color: #ffed4e;
      box-shadow: 0 20px 40px rgba(255, 215, 0, 0.2);
    }
  }
  
  &.disabled {
    opacity: 0.6;
    cursor: not-allowed;
    
    &:hover {
      transform: none;
      box-shadow: none;
    }
  }
}

.vip-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
  color: #2c2c2c;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 4px;
  z-index: 2;
  
  i {
    font-size: 14px;
  }
}

.card-header {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.template-icon {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  
  .icon-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .icon-placeholder {
    color: #9aa0a9;
    font-size: 32px;
  }
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.template-name {
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
  margin: 0;
  line-height: 1.2;
}

.template-desc {
  font-size: 14px;
  color: #b8bcc5;
  line-height: 1.5;
  margin: 0;
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.template-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: auto;
}

.tag {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
  padding: 4px 8px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid rgba(64, 158, 255, 0.3);
}

.card-footer {
  margin-top: 20px;
}

.action-button {
  width: 100%;
  padding: 12px;
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  text-align: center;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  
  &.disabled {
    background: rgba(255, 0, 0, 0.1);
    color: #ff6b6b;
    border-color: rgba(255, 0, 0, 0.3);
  }
}

.hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(45deg, rgba(64, 158, 255, 0.1) 0%, rgba(103, 194, 58, 0.1) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
  border-radius: 16px;
}

// 移动端优化
@media (max-width: 768px) {
  .template-card {
    height: 240px;
    padding: 20px;
  }
  
  .template-icon {
    width: 60px;
    height: 60px;
  }
  
  .template-name {
    font-size: 18px;
  }
  
  .template-desc {
    font-size: 13px;
    -webkit-line-clamp: 2;
  }
  
  .vip-badge {
    top: 8px;
    right: 8px;
    padding: 4px 8px;
    font-size: 11px;
  }
}

@media (max-width: 480px) {
  .template-card {
    height: 200px;
    padding: 16px;
    
    &:hover {
      transform: none;
    }
  }
  
  .template-icon {
    width: 50px;
    height: 50px;
    border-radius: 12px;
  }
  
  .template-name {
    font-size: 16px;
  }
  
  .card-body {
    gap: 8px;
  }
  
  .action-button {
    padding: 10px;
    font-size: 13px;
  }
}
</style>