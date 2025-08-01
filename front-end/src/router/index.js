/*
 * @Author: mangohow mghgyf@qq.com
 * @Date: 2022-12-17 14:35:24
 * @LastEditors: mangohow mghgyf@qq.com
 * @LastEditTime: 2022-12-17 15:22:37
 * @FilePath: \front-end\src\router\index.js
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import Vue from 'vue'
import VueRouter from 'vue-router'

const Login = () => import('../views/Login.vue')
const DashBoard = () => import("../views/DashBoard.vue")
const TemplateView = () => import("../components/TemplateView.vue")
const SpaceView = () => import("../components/SpaceView.vue")
const Payment = () => import("../views/Payment.vue")
const PaymentSuccess = () => import("../views/PaymentSuccess.vue")

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    redirect: "/login"
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/oauth/success',
    name: 'OAuthSuccess',
    component: Login  // 重用登录组件，在mounted中处理OAuth成功逻辑
  },
  {
    path: '/payment',
    name: 'Payment',
    component: Payment
  },
  {
    path: '/payment/success',
    name: 'PaymentSuccess',
    component: PaymentSuccess
  },
  {
    path: '/dash',
    name: 'Dashboard',
    component: DashBoard,
    redirect: "/dash/templates",
    children: [
      {path: "templates", component: TemplateView, meta: {keepAlive: true}},
      {path: "workspaces", component: SpaceView},
      {path: "payment", component: Payment},
      {path: "payment/success", component: PaymentSuccess}
    ]
  }
]

const router = new VueRouter({
  routes
})

//挂载路由导航守卫,此守卫是用来拦截页面访问的，如果没有token，则会被重定向到登录页
// 在登陆时不需要token，直接放行
router.beforeEach((to, from, next) => {
  // to: 将要访问的路径
  // form: 从哪个路径跳转而来
  // next 是一个函数，表示放行
  // next() 放行 next("/login") 强制跳转
  // 如果访问的不是管理员页面，直接放行
  let requestUrl = to.path
  let start = requestUrl.indexOf("/dash")
  if (start != 0) {
    return next();
  }

  // 访问的是登陆后的页面，查看有没有token，如果有，则放行，否则，跳转到登录页
  const tokenStr = window.sessionStorage.getItem("token");
  if(!tokenStr) return next("/login");
  next();
})



export default router
