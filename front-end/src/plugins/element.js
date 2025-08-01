import Vue from 'vue'

import {Container, Header, Aside, Main, Form, FormItem, Button, Message, 
    MessageBox, Input, Dialog, Menu, MenuItem, Select, Option, Tooltip, Loading} from 'element-ui'

Vue.use(Container)
Vue.use(Header)
Vue.use(Aside)
Vue.use(Main)
Vue.use(Form)
Vue.use(Button)      //注册为全局可用的组件
Vue.use(FormItem)
Vue.use(Input)
Vue.use(Dialog)
Vue.use(Menu)
Vue.use(MenuItem)
Vue.use(Select)
Vue.use(Option)
Vue.use(Tooltip)
Vue.use(Loading)

Vue.prototype.$message = Message          //每个Vue对象可用通过this访问Message组件
Vue.prototype.$messageBox = MessageBox