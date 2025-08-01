/*
 * @Author: mangohow mghgyf@qq.com
 * @Date: 2022-12-17 15:38:36
 * @LastEditors: mangohow mghgyf@qq.com
 * @LastEditTime: 2022-12-17 15:38:37
 * @FilePath: \front-end\vue.config.js
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%配%E7%BD%AE
 */

module.exports = {
    publicPath: process.env.NODE_ENV === 'production' ? '/cloud-ide/' : '/',
    devServer: {              //设置本地域名
        host: "0.0.0.0",      // 改为监听所有地址
        port: 8080,
        public: "tiantianai.co",  // 使用生产域名
        disableHostCheck: true,       // 禁用host检查，允许外部访问
        proxy: {          //设置代理解决跨域问题
            "/api": {
                target: "https://tiantianai.co",    // 代理到生产域名
                changeOrigin: true,                 //是否开启跨域
                secure: false                       // 忽略SSL证书验证
            },
            "/auth": {
                target: "https://tiantianai.co",    // 代理到生产域名
                changeOrigin: true,                 //是否开启跨域
                secure: false                       // 忽略SSL证书验证
            }
        }
    }
}