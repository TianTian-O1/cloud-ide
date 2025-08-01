<template>
  <div class="payment-container">
    <div class="payment-header">
      <h1>{{ vipInfo && vipInfo.is_active ? 'VIP管理' : '升级VIP' }}</h1>
      <p>{{ vipInfo && vipInfo.is_active ? '管理您的VIP服务' : '解锁更多功能，享受极致体验' }}</p>
    </div>

    <!-- 当前VIP状态 -->
    <div class="current-status" v-if="vipInfo">
      <div class="status-card" :class="{ vip: vipInfo.is_active }">
        <h3>{{ vipInfo.is_active ? 'VIP用户' : '普通用户' }}</h3>
        <p v-if="vipInfo.is_active && vipInfo.expire_time">
          到期时间：{{ formatDate(vipInfo.expire_time) }}
        </p>
        <p v-if="vipInfo.is_active && vipInfo.current_level">
          当前等级：{{ vipInfo.current_level }}
        </p>
        <p v-if="vipInfo.is_active && vipInfo.days_left">
          剩余天数：{{ vipInfo.days_left }}天
        </p>
        <p v-if="!vipInfo.is_active">升级VIP享受更多特权</p>
      </div>
    </div>

    <!-- VIP套餐选择 -->
    <div class="vip-plans" v-if="!vipInfo || !vipInfo.is_active">
      <h2>选择套餐</h2>
      <div class="plans-grid">
        <div 
          v-for="plan in vipPlans" 
          :key="plan.type"
          class="plan-card"
          :class="{ active: selectedPlan === plan.type }"
          @click="selectPlan(plan)"
        >
          <div class="plan-badge" v-if="plan.type === 'month'">推荐</div>
          <h3>{{ plan.name }}</h3>
          <div class="price">
            <span class="currency">¥</span>
            <span class="amount">{{ plan.price }}</span>
          </div>
          <div class="duration">{{ plan.duration_days }}天</div>
          <ul class="features">
            <li>✓ 无限工作空间</li>
            <li>✓ 高级模板</li>
            <li>✓ 延长运行时间</li>
            <li>✓ 优先技术支持</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 支付方式选择 -->
    <div class="payment-methods" v-if="selectedPlan && (!vipInfo || !vipInfo.is_active)">
      <h2>选择支付方式</h2>
      <div class="methods-grid">
        <div 
          v-for="method in paymentMethods" 
          :key="method.value"
          class="method-card"
          :class="{ active: selectedMethod === method.value }"
          @click="selectedMethod = method.value"
        >
          <div class="method-icon">{{ method.icon }}</div>
          <span>{{ method.name }}</span>
        </div>
      </div>
    </div>

    <!-- 订单确认 -->
    <div class="order-summary" v-if="selectedPlan && selectedMethod && (!vipInfo || !vipInfo.is_active)">
      <h2>订单确认</h2>
      <div class="summary-card">
        <div class="order-item">
          <span>套餐：{{ getCurrentPlan().name }}</span>
          <span>¥{{ getCurrentPlan().price }}</span>
        </div>
        <div class="order-total">
          <span>总计：</span>
          <span class="total-price">¥{{ getCurrentPlan().price }}</span>
        </div>
      </div>
      
      <button 
        class="pay-button" 
        @click="createOrder"
        :disabled="loading"
      >
        {{ loading ? '处理中...' : '立即支付' }}
      </button>
    </div>

    <!-- 支付结果弹窗 -->
    <div class="payment-modal" v-if="showPaymentModal" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>支付信息</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="payment-result">
          <div v-if="paymentResult === 'success'" class="success">
            <div class="success-icon">✓</div>
            <h4>支付成功！</h4>
            <p>VIP权限已激活</p>
          </div>
          <div v-else-if="paymentResult === 'failed'" class="failed">
            <div class="failed-icon">✗</div>
            <h4>支付失败</h4>
            <p>请重试或联系客服</p>
          </div>
          <div v-else-if="paymentResult === 'redirect'" class="redirect">
            <div class="redirect-icon">💳</div>
            <h4>支付订单已创建成功</h4>
            
            <div class="payment-info">
              <p>订单号：{{ currentOrder ? currentOrder.order.order_no : '加载中...' }}</p>
              <p>金额：¥{{ currentOrder ? currentOrder.order.amount : '0.00' }}</p>
              <p>支付方式：{{ getCurrentMethodName() }}</p>
            </div>
            
            <div class="payment-actions">
              <a :href="currentOrder.payUrl" target="_blank" class="pay-btn" @click="handlePaymentClick">
                🚀 立即前往支付
              </a>
              <button @click="copyPaymentLink" class="copy-btn">📋 复制支付链接</button>
            </div>
            
            <p class="pay-tips">点击上方按钮会打开支付页面，请在新页面完成支付操作</p>
          </div>
          <div v-else class="pending">
            <div class="loading">⚡</div>
            <h4>正在处理支付...</h4>
            <p>请稍候，正在连接支付网关</p>
          </div>
        </div>
        <div class="order-info" v-if="currentOrder">
          <p>订单号：{{ currentOrder.order ? currentOrder.order.order_no : 'N/A' }}</p>
          <p>金额：¥{{ currentOrder.order ? currentOrder.order.amount : '0.00' }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Payment',
  data() {
    return {
      vipPlans: [],
      vipInfo: null,
      selectedPlan: null,
      selectedMethod: 'alipay',
      loading: false,
      showPaymentModal: false,
      paymentResult: null, // 'success', 'failed', 'pending'
      currentOrder: null,
      paymentMethods: [
        {
          value: 'alipay',
          name: '支付宝',
          icon: '💰'
        },
        {
          value: 'wechat',
          name: '微信支付',
          icon: '💬'
        },
        {
          value: 'qq',
          name: 'QQ钱包',
          icon: '🐧'
        }
      ]
    }
  },
  mounted() {
    this.loadVipPlans()
    this.loadVipInfo()
    
    // 监听页面可见性变化，当用户返回时重新加载VIP信息
    document.addEventListener('visibilitychange', this.handleVisibilityChange)
  },
  
  beforeDestroy() {
    document.removeEventListener('visibilitychange', this.handleVisibilityChange)
  },
  
  activated() {
    // 当页面被激活时重新加载VIP信息
    this.loadVipInfo()
  },
  methods: {
    async loadVipPlans() {
      try {
        console.log('🔄 开始加载套餐信息...')
        const {data: res} = await this.$axios.get('/api/payment/products')
        console.log('📦 套餐API响应:', res)
        
        if (res.status === 0) {
          this.vipPlans = res.data
          console.log('✅ 套餐信息加载成功:', res.data)
        } else {
          console.error('❌ 套餐API返回错误状态:', res.status, res.message)
          this.$message.error(res.message || '加载套餐信息失败')
        }
      } catch (error) {
        console.error('💥 加载VIP套餐异常:', error)
        console.error('错误详情:', {
          message: error.message,
          response: error.response?.data,
          status: error.response?.status
        })
        this.$message.error('加载套餐信息失败: ' + (error.response?.data?.message || error.message))
      }
    },
    
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
    
    selectPlan(plan) {
      this.selectedPlan = plan.type
    },
    
    getCurrentPlan() {
      return this.vipPlans.find(plan => plan.type === this.selectedPlan)
    },
    
    getCurrentMethodName() {
      const method = this.paymentMethods.find(m => m.value === this.selectedMethod)
      return method ? method.name : ''
    },
    
    async createOrder() {
      if (!this.selectedPlan || !this.selectedMethod) {
        this.$message.warning('请选择套餐和支付方式')
        return
      }
      
      // 获取选中的套餐信息
      const selectedPlanInfo = this.getCurrentPlan()
      if (!selectedPlanInfo) {
        this.$message.error('请选择有效的套餐')
        return
      }
      
      this.loading = true
      this.paymentResult = 'pending'
      this.showPaymentModal = true
      
      try {
        const {data: res} = await this.$axios.post('/api/payment/order', {
          product_id: selectedPlanInfo.id,
          product_type: this.selectedPlan,
          payment_method: this.selectedMethod
        })
        
        if (res.status === 0) {
          this.currentOrder = res.data
          
          // 添加调试信息
          console.log('支付响应数据:', res.data)
          console.log('payUrl存在:', !!res.data.payUrl)
          console.log('qrCode存在:', !!res.data.qrCode)
          console.log('qrCode值:', res.data.qrCode)
          console.log('currentOrder设置为:', res.data)
          
          // 验证数据结构
          if (res.data.qrCode) {
            console.log('✅ 二维码链接有效:', res.data.qrCode)
          } else {
            console.log('❌ 二维码链接缺失')
          }
          
          // 处理支付响应
          if (res.data.payUrl) {
            // 显示支付界面
            this.paymentResult = 'redirect'
            this.currentOrder = res.data
            this.$message.success('支付订单创建成功，请点击按钮前往支付')
            console.log('✅ 支付链接已生成:', res.data.payUrl)
          } else {
            this.$message.error('支付链接生成失败，请重试')
            this.paymentResult = 'failed'
          }
          // 暂时注释掉loadOrders避免干扰
          // this.loadOrders()
        } else {
          this.paymentResult = 'failed'
          this.$message.error(res.message || '创建订单失败')
        }
        
      } catch (error) {
        console.error('创建订单失败:', error)
        this.paymentResult = 'failed'
        this.$message.error('创建订单失败，请重试')
      } finally {
        this.loading = false
      }
    },
    
    closeModal() {
      this.showPaymentModal = false
      this.paymentResult = null
      this.currentOrder = null
    },
    
    handlePaymentClick() {
      // 支付链接点击处理
      const paymentUrl = this.currentOrder.payUrl || this.currentOrder.qrCode
      console.log('用户点击支付链接:', paymentUrl)
      this.$message.info('已打开支付页面，请在新窗口完成支付')
      
      // 可以在这里添加支付状态监听或其他逻辑
      setTimeout(() => {
        this.$message.info('如支付完成，页面将自动更新VIP状态')
      }, 2000)
    },
    

    
    copyPaymentLink() {
      const paymentUrl = this.currentOrder?.payUrl
      if (!paymentUrl) {
        this.$message.warning('没有可复制的支付链接')
        return
      }
      
      if (navigator.clipboard) {
        navigator.clipboard.writeText(paymentUrl).then(() => {
          this.$message.success('支付链接已复制到剪贴板')
        }).catch(() => {
          this.$message.error('复制失败，请手动复制链接')
        })
      } else {
        this.$message.warning('浏览器不支持自动复制，请手动复制支付链接')
      }
    },
    

    
    formatDate(dateStr) {
      if (!dateStr) return ''
      return new Date(dateStr).toLocaleString('zh-CN')
    },
    
    handleVisibilityChange() {
      // 当页面从不可见变为可见时，重新加载VIP信息
      if (!document.hidden) {
        console.log('页面变为可见，重新加载VIP信息')
        this.loadVipInfo()
      }
    }
  }
}
</script>

<style lang="less" scoped>
.payment-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  color: #fff;
}

.payment-header {
  text-align: center;
  margin-bottom: 40px;
  
  h1 {
    font-size: 32px;
    color: #4bc4a5;
    margin-bottom: 10px;
  }
  
  p {
    font-size: 16px;
    color: #aaa;
  }
}

.current-status {
  margin-bottom: 40px;
  
  .status-card {
    background: #2d3139;
    border-radius: 12px;
    padding: 20px;
    text-align: center;
    border: 2px solid #444;
    
    &.vip {
      border-color: #4bc4a5;
      background: linear-gradient(135deg, #2d3139 0%, #1a1d23 100%);
    }
    
    h3 {
      margin-bottom: 10px;
      color: #4bc4a5;
    }
    
    p {
      color: #aaa;
      margin: 5px 0;
    }
  }
}

.vip-plans {
  margin-bottom: 40px;
  
  h2 {
    color: #fff;
    margin-bottom: 20px;
    font-size: 24px;
  }
  
  .plans-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 20px;
  }
  
  .plan-card {
    background: #2d3139;
    border-radius: 12px;
    padding: 24px;
    cursor: pointer;
    position: relative;
    border: 2px solid #444;
    transition: all 0.3s ease;
    
    &:hover {
      border-color: #4bc4a5;
      transform: translateY(-4px);
    }
    
    &.active {
      border-color: #4bc4a5;
      background: linear-gradient(135deg, #2d3139 0%, #1a1d23 100%);
    }
    
    .plan-badge {
      position: absolute;
      top: -10px;
      right: 20px;
      background: #4bc4a5;
      color: #fff;
      padding: 4px 12px;
      border-radius: 12px;
      font-size: 12px;
      font-weight: bold;
    }
    
    h3 {
      color: #4bc4a5;
      margin-bottom: 16px;
      font-size: 20px;
    }
    
    .price {
      margin-bottom: 8px;
      
      .currency {
        font-size: 20px;
        color: #aaa;
      }
      
      .amount {
        font-size: 32px;
        font-weight: bold;
        color: #fff;
        margin-left: 4px;
      }
    }
    
    .duration {
      color: #aaa;
      margin-bottom: 20px;
    }
    
    .features {
      list-style: none;
      padding: 0;
      
      li {
        color: #ccc;
        margin-bottom: 8px;
        font-size: 14px;
      }
    }
  }
}

// 添加支付界面样式
.qrcode {
  text-align: center;
  padding: 20px;
  
  .qrcode-icon {
    font-size: 48px;
    margin-bottom: 16px;
  }
  
  h4 {
    color: #4bc4a5;
    margin-bottom: 20px;
  }
  
  .payment-methods-container {
    margin: 20px 0;
    text-align: center;
  }
  
  .payment-redirect, .qrcode-payment {
    .payment-icon {
      font-size: 48px;
      margin: 20px 0;
    }
    
    .payment-info {
      color: #4bc4a5;
      font-size: 18px;
      font-weight: bold;
      margin: 16px 0;
    }
    
    .main-pay-btn {
      display: inline-block;
      background: linear-gradient(135deg, #1677ff 0%, #69c0ff 100%);
      color: #fff;
      padding: 16px 32px;
      border-radius: 12px;
      text-decoration: none;
      font-weight: bold;
      font-size: 18px;
      margin: 20px 0;
      transition: all 0.3s ease;
      box-shadow: 0 4px 15px rgba(22, 119, 255, 0.3);
      
      &:hover {
        background: linear-gradient(135deg, #0958d9 0%, #40a9ff 100%);
        transform: translateY(-2px);
        box-shadow: 0 6px 20px rgba(22, 119, 255, 0.4);
      }
    }
    
    .payment-actions {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 15px;
      margin: 20px 0;
      flex-wrap: wrap;
      
      .copy-btn, .cancel-btn {
        padding: 10px 20px;
        border: none;
        border-radius: 8px;
        font-weight: bold;
        cursor: pointer;
        transition: all 0.3s ease;
        
        &:hover {
          transform: translateY(-2px);
        }
      }
      
      .copy-btn {
        background: #4bc4a5;
        color: #fff;
        
        &:hover {
          background: #3a9d85;
        }
      }
      
      .cancel-btn {
        background: #6c757d;
        color: #fff;
        
        &:hover {
          background: #5a6268;
        }
      }
      
      .backup-link {
        color: #1677ff;
        text-decoration: none;
        padding: 8px 16px;
        border: 1px solid #1677ff;
        border-radius: 6px;
        transition: all 0.3s ease;
        
        &:hover {
          background: #1677ff;
          color: #fff;
          transform: translateY(-2px);
        }
      }
    }
    
    .payment-tips {
      color: #aaa;
      font-size: 14px;
      margin-top: 20px;
      line-height: 1.5;
    }
  }
  
  .redirect {
    text-align: center;
    padding: 20px;
    
    .redirect-icon {
      font-size: 48px;
      margin-bottom: 15px;
    }
    
    .payment-info {
      background: #1e2128;
      border-radius: 8px;
      padding: 15px;
      margin: 20px 0;
      text-align: left;
      border: 1px solid #444;
      
      p {
        margin: 5px 0;
        color: #ccc;
      }
    }
    
    .payment-actions {
      display: flex;
      gap: 10px;
      justify-content: center;
      margin: 20px 0;
      flex-wrap: wrap;
    }
    
    .pay-btn {
      background: #4bc4a5;
      color: white;
      padding: 12px 24px;
      border: none;
      border-radius: 8px;
      text-decoration: none;
      font-size: 16px;
      cursor: pointer;
      transition: background 0.3s;
      
      &:hover {
        background: #3aa085;
      }
    }
    
    .copy-btn {
      background: #28a745;
      color: white;
      padding: 12px 24px;
      border: none;
      border-radius: 8px;
      font-size: 16px;
      cursor: pointer;
      transition: background 0.3s;
      
      &:hover {
        background: #218838;
      }
    }
    
    .pay-tips {
      font-size: 14px;
      color: #aaa;
      margin-top: 15px;
    }
  }
  
  .qrcode-image {
    margin: 20px 0;
    
    img {
      border: 2px solid #4bc4a5;
      border-radius: 8px;
      background: #fff;
      padding: 10px;
    }
  }
  
  .payment-link {
    margin: 20px 0;
    
    .pay-link-btn {
      display: inline-block;
      background: #4bc4a5;
      color: #fff;
      padding: 12px 24px;
      border-radius: 8px;
      text-decoration: none;
      font-weight: bold;
      transition: background 0.3s ease;
      
      &:hover {
        background: #3a9d85;
      }
    }
  }
  
  .pay-tips {
    color: #aaa;
    font-size: 14px;
    margin-top: 16px;
  }
}

.payment-methods {
  margin-bottom: 40px;
  
  h2 {
    color: #fff;
    margin-bottom: 20px;
    font-size: 24px;
  }
  
  .methods-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 16px;
  }
  
  .method-card {
    background: #2d3139;
    border-radius: 12px;
    padding: 20px;
    text-align: center;
    cursor: pointer;
    border: 2px solid #444;
    transition: all 0.3s ease;
    
    &:hover, &.active {
      border-color: #4bc4a5;
    }
    
    .method-icon {
      font-size: 24px;
      margin-bottom: 8px;
    }
    
    span {
      color: #fff;
      font-weight: 500;
    }
  }
}

.order-summary {
  margin-bottom: 40px;
  
  h2 {
    color: #fff;
    margin-bottom: 20px;
    font-size: 24px;
  }
  
  .summary-card {
    background: #2d3139;
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 20px;
    border: 1px solid #444;
    
    .order-item {
      display: flex;
      justify-content: space-between;
      margin-bottom: 16px;
      color: #ccc;
    }
    
    .order-total {
      display: flex;
      justify-content: space-between;
      font-weight: bold;
      font-size: 18px;
      border-top: 1px solid #444;
      padding-top: 16px;
      
      .total-price {
        color: #4bc4a5;
      }
    }
  }
  
  .pay-button {
    width: 100%;
    background: linear-gradient(135deg, #4bc4a5 0%, #3aa085 100%);
    color: #fff;
    border: none;
    border-radius: 12px;
    padding: 16px;
    font-size: 18px;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 8px 25px rgba(75, 196, 165, 0.3);
    }
    
    &:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }
  }
}

.payment-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  
  .modal-content {
    background: #2d3139;
    border-radius: 12px;
    padding: 30px;
    max-width: 400px;
    width: 90%;
    
    .modal-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      
      h3 {
        color: #4bc4a5;
        margin: 0;
      }
      
      .close-btn {
        background: none;
        border: none;
        color: #aaa;
        font-size: 24px;
        cursor: pointer;
        
        &:hover {
          color: #fff;
        }
      }
    }
    
    .payment-result {
      text-align: center;
      margin-bottom: 20px;
      
      .success-icon, .failed-icon, .loading {
        font-size: 48px;
        margin-bottom: 16px;
      }
      
      .success .success-icon {
        color: #4bc4a5;
      }
      
      .failed .failed-icon {
        color: #ff6b6b;
      }
      
      .loading {
        animation: pulse 1.5s infinite;
        color: #4bc4a5;
      }
      
      h4 {
        color: #fff;
        margin-bottom: 8px;
      }
      
      p {
        color: #aaa;
      }
    }
    
    .order-info {
      background: #1a1d23;
      border-radius: 8px;
      padding: 16px;
      
      p {
        color: #ccc;
        margin: 4px 0;
        font-size: 14px;
      }
    }
  }
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

/* 移动端优化样式 */
@media (max-width: 768px) {
  .payment-container {
    padding: 15px;
    font-size: 14px;
  }
  
  .payment-header {
    text-align: center;
    margin-bottom: 20px;
    
    h1 {
      font-size: 22px;
      margin-bottom: 8px;
    }
    
    p {
      font-size: 13px;
      color: #999;
    }
  }
  
  .current-status {
    .status-card {
      padding: 16px;
      border-radius: 8px;
      
      h3 {
        font-size: 16px;
        margin-bottom: 8px;
      }
      
      p {
        font-size: 13px;
        margin: 4px 0;
      }
    }
  }
  
  .vip-plans {
    h2 {
      font-size: 18px;
      text-align: center;
      margin-bottom: 16px;
    }
    
    .plans-grid {
      display: block;
      
      .plan-card {
        margin-bottom: 16px;
        padding: 16px;
        border-radius: 8px;
        
        h3 {
          font-size: 16px;
          margin-bottom: 10px;
        }
        
        .price {
          margin: 10px 0;
          
          .currency {
            font-size: 14px;
          }
          
          .amount {
            font-size: 20px;
          }
        }
        
        .duration {
          font-size: 12px;
          margin-bottom: 12px;
        }
        
        .features {
          padding-left: 16px;
          
          li {
            font-size: 12px;
            margin-bottom: 4px;
          }
        }
      }
    }
  }
  
  .payment-methods {
    .methods-grid {
      display: block;
      
      .method-item {
        margin-bottom: 12px;
        padding: 12px;
        border-radius: 6px;
        
        .method-icon {
          font-size: 18px;
          margin-right: 8px;
        }
        
        .method-name {
          font-size: 14px;
        }
      }
    }
  }
  
  .payment-actions {
    .pay-button {
      width: 100%;
      padding: 14px;
      font-size: 16px;
      border-radius: 8px;
    }
  }
  
  .payment-modal {
    .modal-content {
      margin: 10px;
      padding: 20px;
      border-radius: 8px;
      
      h3 {
        font-size: 16px;
        margin-bottom: 12px;
      }
      
      .payment-info {
        font-size: 13px;
        
        p {
          margin: 6px 0;
        }
      }
      
      .qr-section {
        text-align: center;
        margin: 16px 0;
        
        img {
          max-width: 180px;
          height: auto;
        }
      }
      
      .close-btn {
        width: 100%;
        padding: 12px;
        margin-top: 16px;
        font-size: 14px;
      }
    }
  }
}
</style> 
 
 
 
 
 