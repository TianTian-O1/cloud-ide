<template>
  <div class="payment-success">
    <div class="success-container">
      <div class="success-icon" v-if="verificationStatus === 'success'">✅</div>
      <div class="loading-icon" v-else-if="verificationStatus === 'loading'">⚡</div>
      <div class="error-icon" v-else-if="verificationStatus === 'error'">❌</div>
      
      <h1 v-if="verificationStatus === 'success'">支付成功！</h1>
      <h1 v-else-if="verificationStatus === 'loading'">正在验证支付...</h1>
      <h1 v-else-if="verificationStatus === 'error'">支付验证失败</h1>
      
      <div class="order-info" v-if="paymentData">
        <p><strong>订单号：</strong>{{ paymentData.out_trade_no }}</p>
        <p><strong>支付金额：</strong>¥{{ paymentData.money }}</p>
        <p><strong>支付方式：</strong>{{ getPaymentTypeName(paymentData.type) }}</p>
        <p><strong>交易号：</strong>{{ paymentData.trade_no }}</p>
        <p><strong>支付时间：</strong>{{ paymentTime }}</p>
        <p><strong>支付状态：</strong><span class="status-success">{{ paymentData.trade_status }}</span></p>
      </div>
      
      <div class="verification-info" v-if="verificationStatus === 'loading'">
        <p>正在验证支付状态并激活VIP权限...</p>
        <div class="loading-bar">
          <div class="loading-progress" :style="{width: progress + '%'}"></div>
        </div>
      </div>
      
      <div class="success-message" v-if="verificationStatus === 'success'">
        <p class="vip-message">🎉 恭喜！您的VIP权限已成功激活</p>
        <p class="activation-time">VIP权限已于 {{ new Date().toLocaleString('zh-CN') }} 激活</p>
      </div>
      
      <div class="error-message" v-if="verificationStatus === 'error'">
        <p>支付验证失败，可能的原因：</p>
        <ul>
          <li>签名验证失败</li>
          <li>订单状态异常</li>
          <li>网络连接问题</li>
        </ul>
        <p>请联系客服处理：support@tiantianai.co</p>
      </div>
      
      <div class="actions">
        <button class="btn-primary" @click="backToPayment" v-if="verificationStatus === 'error'">重新支付</button>
        <button class="btn-primary" @click="backToHome">返回首页</button>
        <button class="btn-secondary" @click="checkVipStatus">查看VIP状态</button>
      </div>
      
      <div class="debug-info" v-if="showDebug">
        <details>
          <summary>调试信息（点击展开）</summary>
          <pre>{{ JSON.stringify(paymentData, null, 2) }}</pre>
          <pre>验证状态: {{ verificationStatus }}</pre>
          <pre>错误信息: {{ errorMessage }}</pre>
        </details>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PaymentSuccess',
  data() {
    return {
      paymentData: null,
      paymentTime: '',
      verificationStatus: 'loading', // 'loading', 'success', 'error'
      progress: 0,
      errorMessage: '',
      showDebug: process.env.NODE_ENV === 'development'
    }
  },
  mounted() {
    this.initPaymentInfo()
    this.startProgress()
    this.verifyPayment()
  },
  methods: {
    initPaymentInfo() {
      // 从hash路由中解析参数（处理 #/payment/success?param=value 格式）
      const hash = window.location.hash
      const queryStart = hash.indexOf('?')
      
      if (queryStart !== -1) {
        const queryString = hash.substring(queryStart + 1)
        const params = new URLSearchParams(queryString)
        
        this.paymentData = {
          pid: params.get('pid'),
          trade_no: params.get('trade_no'),
          out_trade_no: params.get('out_trade_no'),
          type: params.get('type'),
          name: decodeURIComponent(params.get('name') || ''),
          money: params.get('money'),
          trade_status: params.get('trade_status'),
          sign: params.get('sign'),
          sign_type: params.get('sign_type')
        }
        
        this.paymentTime = new Date().toLocaleString('zh-CN')
        
        console.log('解析的支付数据:', this.paymentData)
      } else {
        console.error('未找到支付回调参数')
        this.verificationStatus = 'error'
        this.errorMessage = '未找到支付回调参数'
      }
    },
    
    startProgress() {
      const interval = setInterval(() => {
        if (this.verificationStatus === 'loading') {
          this.progress += 5
          if (this.progress >= 95) {
            this.progress = 95 // 保持在95%直到验证完成
            clearInterval(interval)
          }
        } else {
          this.progress = 100
          clearInterval(interval)
        }
      }, 100)
    },
    
    async verifyPayment() {
      if (!this.paymentData) {
        this.verificationStatus = 'error'
        this.errorMessage = '缺少支付数据'
        return
      }
      
      try {
        console.log('开始验证支付状态...')
        
        // 调用后端API验证支付并处理回调
        const response = await this.$axios.post('/api/payment/callback', {
          pid: this.paymentData.pid,
          trade_no: this.paymentData.trade_no,
          out_trade_no: this.paymentData.out_trade_no,
          type: this.paymentData.type,
          name: this.paymentData.name,
          money: this.paymentData.money,
          trade_status: this.paymentData.trade_status,
          sign: this.paymentData.sign,
          sign_type: this.paymentData.sign_type
        })
        
        console.log('支付验证响应:', response.data)
        
        if (response.data.status === 0) {
          // 验证成功
          this.verificationStatus = 'success'
          this.progress = 100
          this.$message.success('支付验证成功，VIP权限已激活！')
          
          // 延迟3秒后自动跳转到首页
          setTimeout(() => {
            this.$router.push('/')
          }, 3000)
          
        } else {
          throw new Error(response.data.message || '支付验证失败')
        }
        
      } catch (error) {
        console.error('支付验证失败:', error)
        this.verificationStatus = 'error'
        this.errorMessage = error.response?.data?.message || error.message || '网络请求失败'
        this.$message.error('支付验证失败: ' + this.errorMessage)
      }
    },
    
    getPaymentTypeName(type) {
      const typeMap = {
        'alipay': '支付宝',
        'wechat': '微信支付',
        'qq': 'QQ钱包'
      }
      return typeMap[type] || type
    },
    
    backToPayment() {
      this.$router.push('/payment')
    },
    
    backToHome() {
      this.$router.push('/')
    },
    
    checkVipStatus() {
      this.$router.push('/user/vip')
    }
  }
}
</script>

<style lang="less" scoped>
.payment-success {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.success-container {
  background: white;
  border-radius: 20px;
  padding: 40px;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1);
  max-width: 600px;
  width: 100%;
}

.success-icon, .loading-icon, .error-icon {
  font-size: 80px;
  margin-bottom: 20px;
}

.success-icon {
  color: #28a745;
  animation: bounce 1s infinite;
}

.loading-icon {
  color: #007bff;
  animation: pulse 1.5s infinite;
}

.error-icon {
  color: #dc3545;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

h1 {
  color: #2c3e50;
  margin-bottom: 30px;
  font-size: 32px;
  font-weight: bold;
}

.order-info {
  background: #f8f9fa;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  p {
    margin: 8px 0;
    color: #495057;
    
    strong {
      color: #2c3e50;
    }
  }
  
  .status-success {
    color: #28a745;
    font-weight: bold;
  }
}

.verification-info {
  margin-bottom: 30px;
  
  p {
    color: #6c757d;
    margin-bottom: 15px;
  }
}

.success-message {
  background: linear-gradient(135deg, #28a745, #20c997);
  color: white;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  
  .vip-message {
    font-size: 18px;
    font-weight: bold;
    margin-bottom: 10px;
  }
  
  .activation-time {
    font-size: 14px;
    opacity: 0.9;
  }
}

.error-message {
  background: #f8d7da;
  color: #721c24;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  ul {
    margin: 10px 0;
    padding-left: 20px;
  }
}

.actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 30px;
  flex-wrap: wrap;
  
  button {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &.btn-primary {
      background: #007bff;
      color: white;
      
      &:hover {
        background: #0056b3;
      }
    }
    
    &.btn-secondary {
      background: #6c757d;
      color: white;
      
      &:hover {
        background: #545b62;
      }
    }
  }
}

.loading-bar {
  width: 100%;
  height: 6px;
  background: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
  
  .loading-progress {
    height: 100%;
    background: linear-gradient(90deg, #007bff, #20c997);
    transition: width 0.3s ease;
    border-radius: 3px;
  }
}

.debug-info {
  margin-top: 30px;
  text-align: left;
  
  details {
    background: #f8f9fa;
    border-radius: 5px;
    padding: 10px;
    
    summary {
      cursor: pointer;
      color: #6c757d;
      font-size: 12px;
    }
    
    pre {
      background: #fff;
      padding: 10px;
      border-radius: 3px;
      font-size: 11px;
      overflow: auto;
      margin: 5px 0;
    }
  }
}
</style> 
</style> 
  <div class="payment-success">
    <div class="success-container">
      <div class="success-icon" v-if="verificationStatus === 'success'">✅</div>
      <div class="loading-icon" v-else-if="verificationStatus === 'loading'">⚡</div>
      <div class="error-icon" v-else-if="verificationStatus === 'error'">❌</div>
      
      <h1 v-if="verificationStatus === 'success'">支付成功！</h1>
      <h1 v-else-if="verificationStatus === 'loading'">正在验证支付...</h1>
      <h1 v-else-if="verificationStatus === 'error'">支付验证失败</h1>
      
      <div class="order-info" v-if="paymentData">
        <p><strong>订单号：</strong>{{ paymentData.out_trade_no }}</p>
        <p><strong>支付金额：</strong>¥{{ paymentData.money }}</p>
        <p><strong>支付方式：</strong>{{ getPaymentTypeName(paymentData.type) }}</p>
        <p><strong>交易号：</strong>{{ paymentData.trade_no }}</p>
        <p><strong>支付时间：</strong>{{ paymentTime }}</p>
        <p><strong>支付状态：</strong><span class="status-success">{{ paymentData.trade_status }}</span></p>
      </div>
      
      <div class="verification-info" v-if="verificationStatus === 'loading'">
        <p>正在验证支付状态并激活VIP权限...</p>
        <div class="loading-bar">
          <div class="loading-progress" :style="{width: progress + '%'}"></div>
        </div>
      </div>
      
      <div class="success-message" v-if="verificationStatus === 'success'">
        <p class="vip-message">🎉 恭喜！您的VIP权限已成功激活</p>
        <p class="activation-time">VIP权限已于 {{ new Date().toLocaleString('zh-CN') }} 激活</p>
      </div>
      
      <div class="error-message" v-if="verificationStatus === 'error'">
        <p>支付验证失败，可能的原因：</p>
        <ul>
          <li>签名验证失败</li>
          <li>订单状态异常</li>
          <li>网络连接问题</li>
        </ul>
        <p>请联系客服处理：support@tiantianai.co</p>
      </div>
      
      <div class="actions">
        <button class="btn-primary" @click="backToPayment" v-if="verificationStatus === 'error'">重新支付</button>
        <button class="btn-primary" @click="backToHome">返回首页</button>
        <button class="btn-secondary" @click="checkVipStatus">查看VIP状态</button>
      </div>
      
      <div class="debug-info" v-if="showDebug">
        <details>
          <summary>调试信息（点击展开）</summary>
          <pre>{{ JSON.stringify(paymentData, null, 2) }}</pre>
          <pre>验证状态: {{ verificationStatus }}</pre>
          <pre>错误信息: {{ errorMessage }}</pre>
        </details>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PaymentSuccess',
  data() {
    return {
      paymentData: null,
      paymentTime: '',
      verificationStatus: 'loading', // 'loading', 'success', 'error'
      progress: 0,
      errorMessage: '',
      showDebug: process.env.NODE_ENV === 'development'
    }
  },
  mounted() {
    this.initPaymentInfo()
    this.startProgress()
    this.verifyPayment()
  },
  methods: {
    initPaymentInfo() {
      // 从hash路由中解析参数（处理 #/payment/success?param=value 格式）
      const hash = window.location.hash
      const queryStart = hash.indexOf('?')
      
      if (queryStart !== -1) {
        const queryString = hash.substring(queryStart + 1)
        const params = new URLSearchParams(queryString)
        
        this.paymentData = {
          pid: params.get('pid'),
          trade_no: params.get('trade_no'),
          out_trade_no: params.get('out_trade_no'),
          type: params.get('type'),
          name: decodeURIComponent(params.get('name') || ''),
          money: params.get('money'),
          trade_status: params.get('trade_status'),
          sign: params.get('sign'),
          sign_type: params.get('sign_type')
        }
        
        this.paymentTime = new Date().toLocaleString('zh-CN')
        
        console.log('解析的支付数据:', this.paymentData)
      } else {
        console.error('未找到支付回调参数')
        this.verificationStatus = 'error'
        this.errorMessage = '未找到支付回调参数'
      }
    },
    
    startProgress() {
      const interval = setInterval(() => {
        if (this.verificationStatus === 'loading') {
          this.progress += 5
          if (this.progress >= 95) {
            this.progress = 95 // 保持在95%直到验证完成
            clearInterval(interval)
          }
        } else {
          this.progress = 100
          clearInterval(interval)
        }
      }, 100)
    },
    
    async verifyPayment() {
      if (!this.paymentData) {
        this.verificationStatus = 'error'
        this.errorMessage = '缺少支付数据'
        return
      }
      
      try {
        console.log('开始验证支付状态...')
        
        // 调用后端API验证支付并处理回调
        const response = await this.$axios.post('/api/payment/callback', {
          pid: this.paymentData.pid,
          trade_no: this.paymentData.trade_no,
          out_trade_no: this.paymentData.out_trade_no,
          type: this.paymentData.type,
          name: this.paymentData.name,
          money: this.paymentData.money,
          trade_status: this.paymentData.trade_status,
          sign: this.paymentData.sign,
          sign_type: this.paymentData.sign_type
        })
        
        console.log('支付验证响应:', response.data)
        
        if (response.data.status === 0) {
          // 验证成功
          this.verificationStatus = 'success'
          this.progress = 100
          this.$message.success('支付验证成功，VIP权限已激活！')
          
          // 延迟3秒后自动跳转到首页
          setTimeout(() => {
            this.$router.push('/')
          }, 3000)
          
        } else {
          throw new Error(response.data.message || '支付验证失败')
        }
        
      } catch (error) {
        console.error('支付验证失败:', error)
        this.verificationStatus = 'error'
        this.errorMessage = error.response?.data?.message || error.message || '网络请求失败'
        this.$message.error('支付验证失败: ' + this.errorMessage)
      }
    },
    
    getPaymentTypeName(type) {
      const typeMap = {
        'alipay': '支付宝',
        'wechat': '微信支付',
        'qq': 'QQ钱包'
      }
      return typeMap[type] || type
    },
    
    backToPayment() {
      this.$router.push('/payment')
    },
    
    backToHome() {
      this.$router.push('/')
    },
    
    checkVipStatus() {
      this.$router.push('/user/vip')
    }
  }
}
</script>

<style lang="less" scoped>
.payment-success {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.success-container {
  background: white;
  border-radius: 20px;
  padding: 40px;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1);
  max-width: 600px;
  width: 100%;
}

.success-icon, .loading-icon, .error-icon {
  font-size: 80px;
  margin-bottom: 20px;
}

.success-icon {
  color: #28a745;
  animation: bounce 1s infinite;
}

.loading-icon {
  color: #007bff;
  animation: pulse 1.5s infinite;
}

.error-icon {
  color: #dc3545;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

h1 {
  color: #2c3e50;
  margin-bottom: 30px;
  font-size: 32px;
  font-weight: bold;
}

.order-info {
  background: #f8f9fa;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  p {
    margin: 8px 0;
    color: #495057;
    
    strong {
      color: #2c3e50;
    }
  }
  
  .status-success {
    color: #28a745;
    font-weight: bold;
  }
}

.verification-info {
  margin-bottom: 30px;
  
  p {
    color: #6c757d;
    margin-bottom: 15px;
  }
}

.success-message {
  background: linear-gradient(135deg, #28a745, #20c997);
  color: white;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  
  .vip-message {
    font-size: 18px;
    font-weight: bold;
    margin-bottom: 10px;
  }
  
  .activation-time {
    font-size: 14px;
    opacity: 0.9;
  }
}

.error-message {
  background: #f8d7da;
  color: #721c24;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  ul {
    margin: 10px 0;
    padding-left: 20px;
  }
}

.actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 30px;
  flex-wrap: wrap;
  
  button {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &.btn-primary {
      background: #007bff;
      color: white;
      
      &:hover {
        background: #0056b3;
      }
    }
    
    &.btn-secondary {
      background: #6c757d;
      color: white;
      
      &:hover {
        background: #545b62;
      }
    }
  }
}

.loading-bar {
  width: 100%;
  height: 6px;
  background: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
  
  .loading-progress {
    height: 100%;
    background: linear-gradient(90deg, #007bff, #20c997);
    transition: width 0.3s ease;
    border-radius: 3px;
  }
}

.debug-info {
  margin-top: 30px;
  text-align: left;
  
  details {
    background: #f8f9fa;
    border-radius: 5px;
    padding: 10px;
    
    summary {
      cursor: pointer;
      color: #6c757d;
      font-size: 12px;
    }
    
    pre {
      background: #fff;
      padding: 10px;
      border-radius: 3px;
      font-size: 11px;
      overflow: auto;
      margin: 5px 0;
    }
  }
}
</style> 
</style> 
  <div class="payment-success">
    <div class="success-container">
      <div class="success-icon" v-if="verificationStatus === 'success'">✅</div>
      <div class="loading-icon" v-else-if="verificationStatus === 'loading'">⚡</div>
      <div class="error-icon" v-else-if="verificationStatus === 'error'">❌</div>
      
      <h1 v-if="verificationStatus === 'success'">支付成功！</h1>
      <h1 v-else-if="verificationStatus === 'loading'">正在验证支付...</h1>
      <h1 v-else-if="verificationStatus === 'error'">支付验证失败</h1>
      
      <div class="order-info" v-if="paymentData">
        <p><strong>订单号：</strong>{{ paymentData.out_trade_no }}</p>
        <p><strong>支付金额：</strong>¥{{ paymentData.money }}</p>
        <p><strong>支付方式：</strong>{{ getPaymentTypeName(paymentData.type) }}</p>
        <p><strong>交易号：</strong>{{ paymentData.trade_no }}</p>
        <p><strong>支付时间：</strong>{{ paymentTime }}</p>
        <p><strong>支付状态：</strong><span class="status-success">{{ paymentData.trade_status }}</span></p>
      </div>
      
      <div class="verification-info" v-if="verificationStatus === 'loading'">
        <p>正在验证支付状态并激活VIP权限...</p>
        <div class="loading-bar">
          <div class="loading-progress" :style="{width: progress + '%'}"></div>
        </div>
      </div>
      
      <div class="success-message" v-if="verificationStatus === 'success'">
        <p class="vip-message">🎉 恭喜！您的VIP权限已成功激活</p>
        <p class="activation-time">VIP权限已于 {{ new Date().toLocaleString('zh-CN') }} 激活</p>
      </div>
      
      <div class="error-message" v-if="verificationStatus === 'error'">
        <p>支付验证失败，可能的原因：</p>
        <ul>
          <li>签名验证失败</li>
          <li>订单状态异常</li>
          <li>网络连接问题</li>
        </ul>
        <p>请联系客服处理：support@tiantianai.co</p>
      </div>
      
      <div class="actions">
        <button class="btn-primary" @click="backToPayment" v-if="verificationStatus === 'error'">重新支付</button>
        <button class="btn-primary" @click="backToHome">返回首页</button>
        <button class="btn-secondary" @click="checkVipStatus">查看VIP状态</button>
      </div>
      
      <div class="debug-info" v-if="showDebug">
        <details>
          <summary>调试信息（点击展开）</summary>
          <pre>{{ JSON.stringify(paymentData, null, 2) }}</pre>
          <pre>验证状态: {{ verificationStatus }}</pre>
          <pre>错误信息: {{ errorMessage }}</pre>
        </details>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PaymentSuccess',
  data() {
    return {
      paymentData: null,
      paymentTime: '',
      verificationStatus: 'loading', // 'loading', 'success', 'error'
      progress: 0,
      errorMessage: '',
      showDebug: process.env.NODE_ENV === 'development'
    }
  },
  mounted() {
    this.initPaymentInfo()
    this.startProgress()
    this.verifyPayment()
  },
  methods: {
    initPaymentInfo() {
      // 从hash路由中解析参数（处理 #/payment/success?param=value 格式）
      const hash = window.location.hash
      const queryStart = hash.indexOf('?')
      
      if (queryStart !== -1) {
        const queryString = hash.substring(queryStart + 1)
        const params = new URLSearchParams(queryString)
        
        this.paymentData = {
          pid: params.get('pid'),
          trade_no: params.get('trade_no'),
          out_trade_no: params.get('out_trade_no'),
          type: params.get('type'),
          name: decodeURIComponent(params.get('name') || ''),
          money: params.get('money'),
          trade_status: params.get('trade_status'),
          sign: params.get('sign'),
          sign_type: params.get('sign_type')
        }
        
        this.paymentTime = new Date().toLocaleString('zh-CN')
        
        console.log('解析的支付数据:', this.paymentData)
      } else {
        console.error('未找到支付回调参数')
        this.verificationStatus = 'error'
        this.errorMessage = '未找到支付回调参数'
      }
    },
    
    startProgress() {
      const interval = setInterval(() => {
        if (this.verificationStatus === 'loading') {
          this.progress += 5
          if (this.progress >= 95) {
            this.progress = 95 // 保持在95%直到验证完成
            clearInterval(interval)
          }
        } else {
          this.progress = 100
          clearInterval(interval)
        }
      }, 100)
    },
    
    async verifyPayment() {
      if (!this.paymentData) {
        this.verificationStatus = 'error'
        this.errorMessage = '缺少支付数据'
        return
      }
      
      try {
        console.log('开始验证支付状态...')
        
        // 调用后端API验证支付并处理回调
        const response = await this.$axios.post('/api/payment/callback', {
          pid: this.paymentData.pid,
          trade_no: this.paymentData.trade_no,
          out_trade_no: this.paymentData.out_trade_no,
          type: this.paymentData.type,
          name: this.paymentData.name,
          money: this.paymentData.money,
          trade_status: this.paymentData.trade_status,
          sign: this.paymentData.sign,
          sign_type: this.paymentData.sign_type
        })
        
        console.log('支付验证响应:', response.data)
        
        if (response.data.status === 0) {
          // 验证成功
          this.verificationStatus = 'success'
          this.progress = 100
          this.$message.success('支付验证成功，VIP权限已激活！')
          
          // 延迟3秒后自动跳转到首页
          setTimeout(() => {
            this.$router.push('/')
          }, 3000)
          
        } else {
          throw new Error(response.data.message || '支付验证失败')
        }
        
      } catch (error) {
        console.error('支付验证失败:', error)
        this.verificationStatus = 'error'
        this.errorMessage = error.response?.data?.message || error.message || '网络请求失败'
        this.$message.error('支付验证失败: ' + this.errorMessage)
      }
    },
    
    getPaymentTypeName(type) {
      const typeMap = {
        'alipay': '支付宝',
        'wechat': '微信支付',
        'qq': 'QQ钱包'
      }
      return typeMap[type] || type
    },
    
    backToPayment() {
      this.$router.push('/payment')
    },
    
    backToHome() {
      this.$router.push('/')
    },
    
    checkVipStatus() {
      this.$router.push('/user/vip')
    }
  }
}
</script>

<style lang="less" scoped>
.payment-success {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.success-container {
  background: white;
  border-radius: 20px;
  padding: 40px;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.1);
  max-width: 600px;
  width: 100%;
}

.success-icon, .loading-icon, .error-icon {
  font-size: 80px;
  margin-bottom: 20px;
}

.success-icon {
  color: #28a745;
  animation: bounce 1s infinite;
}

.loading-icon {
  color: #007bff;
  animation: pulse 1.5s infinite;
}

.error-icon {
  color: #dc3545;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

@keyframes pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

h1 {
  color: #2c3e50;
  margin-bottom: 30px;
  font-size: 32px;
  font-weight: bold;
}

.order-info {
  background: #f8f9fa;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  p {
    margin: 8px 0;
    color: #495057;
    
    strong {
      color: #2c3e50;
    }
  }
  
  .status-success {
    color: #28a745;
    font-weight: bold;
  }
}

.verification-info {
  margin-bottom: 30px;
  
  p {
    color: #6c757d;
    margin-bottom: 15px;
  }
}

.success-message {
  background: linear-gradient(135deg, #28a745, #20c997);
  color: white;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  
  .vip-message {
    font-size: 18px;
    font-weight: bold;
    margin-bottom: 10px;
  }
  
  .activation-time {
    font-size: 14px;
    opacity: 0.9;
  }
}

.error-message {
  background: #f8d7da;
  color: #721c24;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  text-align: left;
  
  ul {
    margin: 10px 0;
    padding-left: 20px;
  }
}

.actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 30px;
  flex-wrap: wrap;
  
  button {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &.btn-primary {
      background: #007bff;
      color: white;
      
      &:hover {
        background: #0056b3;
      }
    }
    
    &.btn-secondary {
      background: #6c757d;
      color: white;
      
      &:hover {
        background: #545b62;
      }
    }
  }
}

.loading-bar {
  width: 100%;
  height: 6px;
  background: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
  
  .loading-progress {
    height: 100%;
    background: linear-gradient(90deg, #007bff, #20c997);
    transition: width 0.3s ease;
    border-radius: 3px;
  }
}

.debug-info {
  margin-top: 30px;
  text-align: left;
  
  details {
    background: #f8f9fa;
    border-radius: 5px;
    padding: 10px;
    
    summary {
      cursor: pointer;
      color: #6c757d;
      font-size: 12px;
    }
    
    pre {
      background: #fff;
      padding: 10px;
      border-radius: 3px;
      font-size: 11px;
      overflow: auto;
      margin: 5px 0;
    }
  }
}
</style> 
</style> 
 
 
 
 
 