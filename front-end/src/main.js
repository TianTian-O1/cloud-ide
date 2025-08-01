import Vue from 'vue'
import App from './App.vue'
import router from './router'
import axios from "axios"
import './plugins/element.js'
import 'element-ui/lib/theme-chalk/index.css'
import 'element-ui/lib/theme-chalk/display.css'
import './assets/mobile-responsive.css'
import moment from 'moment/moment'
import { Message } from 'element-ui'

Vue.config.productionTip = false

switch (process.env.NODE_ENV) {
  case 'development':
    axios.defaults.baseURL = ""  // 使用nginx代理，相对路径
    axios.defaults.workspaceUrl = "http://localhost:8080/ws/"
    break
  default:
    // 生产环境通过nginx代理
    axios.defaults.baseURL = ""  // 使用nginx代理，相对路径
    axios.defaults.workspaceUrl = "https://tiantianai.co/ws/"
}


Vue.prototype.$axios = axios

//配置请求拦截器，用于在访问后端服务器时携带token令牌
axios.interceptors.request.use(config =>{
  let requestUrl = config.url
  let start = requestUrl.indexOf("/api")
  if (requestUrl.startsWith("/api")) {   // 如果访问后台，需要带上token
    const token = window.sessionStorage.getItem("token")
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }

  return config           //必须return config
})

// 配置响应拦截器，该拦截器用于拦截后端数据的响应
axios.interceptors.response.use(config => {
  return config;
}, error => {
  if (error.response) {
    switch (error.response.status) {
      case 401:
        // 返回 401 清除 token 信息并跳转到登录页面
        window.sessionStorage.clear()
        router.push("/login")
        Message.error("需要登录")
        break
      case 400:
        Message.error("请求参数有误")
        break
    }
  }
  return Promise.reject(error)  // 返回完整error对象而不只是data
})

Vue.filter('dateFormat', function (dateStr) {
  // 根据传入的日期字符串进行格式化
  return moment(dateStr).format('YYYY-MM-DD HH:mm:ss');
});


new Vue({
  router,
  render: h => h(App)
}).$mount('#app')
