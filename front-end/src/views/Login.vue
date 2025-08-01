<template>

  <div class="login_container">
      <div class="login_box">
          <div class="login_font">Cloud Code</div>
          <!-- 登陆区域 -->
          <el-form ref="loginFormRef" :model="loginForm" :rules="loginFormRules" label-width="0px" class="login_form">
              <!-- 用户名 -->
              <el-form-item prop="username">
                  <el-input v-model="loginForm.username" prefix-icon="el-icon-user-solid" placeholder="请输入用户名"></el-input>
              </el-form-item>
              <!-- 密码 -->
              <el-form-item prop="password">
                  <el-input v-model="loginForm.password" @keyup.enter.native="login" prefix-icon="el-icon-lock" type="password" placeholder="请输入密码"></el-input>
              </el-form-item>
              <!-- 按钮区域 -->
              <el-form-item class="btns">
                  <el-button type="primary" @click="login">登录</el-button>
                  <el-button type="primary" @click="resetLoginForm">重置</el-button>
                  <el-button type="primary" @click="showRegisterDialog">注册</el-button>
              </el-form-item>
              
              <!-- OAuth登录区域 -->
              <el-form-item class="oauth-login" v-if="oauthEnabled">
                  <div class="oauth-divider">
                      <span>或</span>
                  </div>
                            <el-button class="linuxdo-login-btn" @click="loginWithLinuxDo" :loading="linuxdoLoading">
            <img src="static/images/linuxdo-logo.png" class="oauth-logo" alt="LinuxDo" />
            使用 LinuxDo 登录
          </el-button>
              </el-form-item>
              
              <!-- 忘记密码链接 -->
              <el-form-item class="forgot-password-link">
                  <el-link type="primary" @click="showForgotPasswordDialog">忘记密码？</el-link>
              </el-form-item>
          </el-form>
      </div>
      
      <!-- 注册dialog -->
      <el-dialog title="用户注册" :visible.sync="dialogFormVisible">

        <el-form ref="registerFormRef" :model="registerForm" :rules="registerFormRules" label-width="0px" label-position="right" class="register_form">
          <el-form-item prop="nickname" label="昵称" label-width="100px">
            <el-input v-model="registerForm.nickname"></el-input>
          </el-form-item>
          <el-form-item prop="username" label="用户名" label-width="100px">
            <el-input v-model="registerForm.username"></el-input>
          </el-form-item>
          <el-form-item prop="password" label="密码" label-width="100px">
            <el-input v-model="registerForm.password" type="password"></el-input>
          </el-form-item>
          <el-form-item prop="email" label="邮箱" label-width="100px" class="form-email">
            <el-input v-model="registerForm.email"></el-input>
            <el-button @click="getEmailCode" :disabled="getEmailCodeButtonEnable" class="vcb">{{ getEmailCodeButtonName }}</el-button>
          </el-form-item>
          <el-form-item prop="emailCode" label="验证码" label-width="100px">
            <el-input v-model="registerForm.emailCode" placeholder="请输入邮箱验证码"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogFormVisible = false">取 消</el-button>
          <el-button type="primary" @click="register">提 交</el-button>
        </div>

      </el-dialog>

      <!-- 忘记密码dialog -->
      <el-dialog title="忘记密码" :visible.sync="forgotPasswordDialogVisible" width="450px">
        <el-form ref="forgotPasswordFormRef" :model="forgotPasswordForm" :rules="forgotPasswordFormRules" label-width="80px" class="forgot-password-form">
          <el-form-item prop="email" label="邮箱">
            <el-input v-model="forgotPasswordForm.email" placeholder="请输入注册时使用的邮箱"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="sendResetCode" :disabled="sendResetCodeButtonEnable" style="width: 100%;">
              {{ sendResetCodeButtonName }}
            </el-button>
          </el-form-item>
          
          <el-form-item prop="emailCode" label="验证码" v-if="showResetForm">
            <el-input v-model="forgotPasswordForm.emailCode" placeholder="请输入邮箱验证码"></el-input>
          </el-form-item>
          <el-form-item prop="newPassword" label="新密码" v-if="showResetForm">
            <el-input v-model="forgotPasswordForm.newPassword" type="password" placeholder="请输入新密码"></el-input>
          </el-form-item>
          <el-form-item prop="confirmPassword" label="确认密码" v-if="showResetForm">
            <el-input v-model="forgotPasswordForm.confirmPassword" type="password" placeholder="请再次输入新密码"></el-input>
          </el-form-item>
        </el-form>
        
        <div slot="footer" class="dialog-footer">
          <el-button @click="closeForgotPasswordDialog">取 消</el-button>
          <el-button type="primary" @click="resetPassword" v-if="showResetForm">重置密码</el-button>
        </div>
      </el-dialog>

  </div>

</template>



<script>
import { json } from 'body-parser'
import md5 from 'js-md5'
import {Base64} from "js-base64"

export default {
  data() {
    const validateUsername = async (rule, value, callback) => {
      const reg = /^\w+$/
      if (!reg.test(value)) {
        callback(new Error("用户名只能包含数字,英文字母和下划线"))
        return
      }
      const url = "/auth/username/check" + "?username=" + value
      const {data: res} = await this.$axios.get(url)
      console.log(res)
      if (res.status === 18) {  // 只有用户名不可用时才报错
        callback(new Error(res.message))
        return
      }
      callback()
    };
    
    const validateEmail = (rule, value, callback) => {
      const regEmail = /^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(\.[a-zA-Z0-9_-])+/
      if (!regEmail.test(value)) {
        callback(new Error('请输入合法的邮箱'))
      } else {
        callback()
      }
    };
    
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.forgotPasswordForm.newPassword) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    };
    
    return {
          // 登录表单的数据绑定对象
          loginForm: {
              username: "",
              password: ""
          },
          //表单的验证规则对象
          loginFormRules: {
              // 验证用户名
              username: [
                  { required: true, message: "请输入用户名", trigger: "blur"},
                  { min: 3, max: 11, message: "长度在 3 到 10 个字符之间", trigger: "blur"}
              ],
              // 验证密码
              password: [
                  { required: true, message: "请输入密码", trigger: "blur"},
                  { min: 4, max: 15, message: "长度在 4 到 15 个字符之间", trigger: "blur"}
              ]
          },
          registerFormRules: {
            nickname: [
              {required: true, message: "请输入昵称", trigger: "blur"},
              {min: 5, max: 32, message: "昵称太短或太长", trigger: "blur"}
            ],
            // 验证用户名
            username: [
                { required: true, message: "请输入用户名", trigger: "blur"},
                { min: 3, max: 10, message: "长度在 3 到 10 个字符之间", trigger: "blur"},
                { validator: validateUsername, trigger: "blur"}
            ],
            // 验证密码
            password: [
                { required: true, message: "请输入密码", trigger: "blur"},
                { min: 8, max: 24, message: "长度在 8 到 24 个字符之间", trigger: "blur"}
            ],
            email: [
                { required: true, message: "请输入邮箱", trigger: "blur"},
                { validator: validateEmail, trigger: "blur"}
            ],
            emailCode: [
                { required: true, message: "请输入验证码", trigger: "blur"},
                { min: 6, max: 6, message: "长度为6", trigger: "blur"}
            ]

          },
          dialogFormVisible: false,
          registerForm: {
            nickname: "",
            username: "",
            password: "",
            email: "",
            emailCode: ""
          },
          getEmailCodeButtonName: "发送验证码",
          getEmailCodeButtonEnable: false,
          countDown: 60,
          
          // 忘记密码相关数据
          forgotPasswordDialogVisible: false,
          showResetForm: false,
          forgotPasswordForm: {
            email: "",
            emailCode: "",
            newPassword: "",
            confirmPassword: ""
          },
          forgotPasswordFormRules: {
            email: [
              { required: true, message: "请输入邮箱", trigger: "blur"},
              { validator: validateEmail, trigger: "blur"}
            ],
            emailCode: [
              { required: true, message: "请输入验证码", trigger: "blur"},
              { min: 6, max: 6, message: "长度为6", trigger: "blur"}
            ],
            newPassword: [
              { required: true, message: "请输入新密码", trigger: "blur"},
              { min: 8, max: 24, message: "长度在 8 到 24 个字符之间", trigger: "blur"}
            ],
            confirmPassword: [
              { required: true, message: "请再次输入新密码", trigger: "blur"},
              { validator: validateConfirmPassword, trigger: "blur"}
            ]
          },
          sendResetCodeButtonName: "发送重置验证码",
          sendResetCodeButtonEnable: false,
          resetCountDown: 60,
          
          // OAuth相关数据
          oauthEnabled: false,
          linuxdoLoading: false,
      }
  },
  methods: {
      resetLoginForm() {
          this.$refs.loginFormRef.resetFields();
      },
      login() {   //表单预校验
          this.$refs.loginFormRef.validate(async (valid) =>{
              if(!valid) return;     //预验证没有通过
              const forms = {
                  username: this.loginForm.username,
                  password: this.loginForm.password
              }
              const {data: res} = await this.$axios.post("/auth/login", forms);
              if(res.status !== 2 && res.status !== 0 && res.status !== 12) {   //登录失败，兼容多种成功状态码
                  return this.$message.error(res.message);
              }
              if (!res.data) {
                  return this.$message.error(res.message);
              }
              const jsonData = JSON.stringify(res.data)
              const encodedData = Base64.encode(jsonData)
              
              // 登录成功之后:
                  //1 将登陆成功的Token保存到客户端的sessionStorage中，token只应在当前网站
                  //    打开期间生效，所以将Token保存到客户端的sessionStorage中
                  // sessionStorage 是会话期间的存储 localStorage是持久化的存储
                  //2 通过编程式导航跳转到后台主页.路由地址为home
              window.sessionStorage.setItem("userData", encodedData)
              window.sessionStorage.setItem("token", res.data.token)
              window.sessionStorage.setItem("userId", res.data.id)
              await this.$router.push("/dash")
          });
      },
      showRegisterDialog() {
        if (this.dialogFormVisible == false) {
          this.dialogFormVisible = true
        }
      },
      // 用户注册
      register() {
        this.$refs.registerFormRef.validate(async (valid) =>{
              if(!valid) {
                this.$message.warning("请检查输入信息：昵称5-32字符，用户名3-10字符，密码8-24字符，邮箱格式正确，验证码6位")
                return;     //预验证没有通过
              }
              const registerForm = {...this.registerForm}
              
              const {data: res} = await this.$axios.post("/auth/register", registerForm)
              if (res.status) {
                this.$message.error(res.message)
                return
              }
              
              this.$message.success(res.message)
              this.dialogFormVisible = false
              this.loginForm.username = this.registerForm.username
              this.registerForm = {nickname: "", username: "", password: "", email: "", emailCode: ""}
          });
      },
      // 获取邮箱验证码
      async getEmailCode() {
        const regEmail = /^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(\.[a-zA-Z0-9_-])+/
        if (!regEmail.test(this.registerForm.email)) {
          this.$message.error('请输入合法的邮箱')
          return
        }

        const url = "/auth/emailCode?email=" + this.registerForm.email
        const {data: res} = await this.$axios.get(url)
        if (res.status) {
          this.$message.error(res.message)
          return
        } else {
          this.$message.success(res.message)
        }

        this.getEmailCodeButtonEnable = true

        var timer = setInterval(() => {
          this.countDown -= 1
          this.getEmailCodeButtonName = this.countDown + "秒后重新发送"
          if (this.countDown == 0) {
            this.getEmailCodeButtonEnable = false
            this.getEmailCodeButtonName = "发送验证码"
            this.countDown = 60
              clearInterval(timer)
            }
        }, 1000)

      },
      
      // 忘记密码相关方法
      showForgotPasswordDialog() {
        this.forgotPasswordDialogVisible = true;
        this.showResetForm = false;
        this.resetForgotPasswordForm();
      },
      
      closeForgotPasswordDialog() {
        this.forgotPasswordDialogVisible = false;
        this.showResetForm = false;
        this.resetForgotPasswordForm();
        if (this.resetTimer) {
          clearInterval(this.resetTimer);
        }
        this.sendResetCodeButtonEnable = false;
        this.sendResetCodeButtonName = "发送重置验证码";
        this.resetCountDown = 60;
      },
      
      resetForgotPasswordForm() {
        this.forgotPasswordForm = {
          email: "",
          emailCode: "",
          newPassword: "",
          confirmPassword: ""
        };
        if (this.$refs.forgotPasswordFormRef) {
          this.$refs.forgotPasswordFormRef.resetFields();
        }
      },
      
      async sendResetCode() {
        // 验证邮箱格式
        if (!this.forgotPasswordForm.email) {
          return this.$message.error("请输入邮箱");
        }
        
        const regEmail = /^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(\.[a-zA-Z0-9_-])+/;
        if (!regEmail.test(this.forgotPasswordForm.email)) {
          return this.$message.error("请输入正确的邮箱格式");
        }
        
        try {
          const {data: res} = await this.$axios.post("/auth/forgot-password", {
            email: this.forgotPasswordForm.email
          });
          
          if (res.status === 0) {
            this.$message.success("验证码发送成功，请查收邮件");
            this.showResetForm = true;
            this.startResetCountDown();
          } else {
            this.$message.error(res.message || "发送失败，请重试");
          }
        } catch (error) {
          this.$message.error("发送失败，请检查网络连接");
        }
      },
      
      startResetCountDown() {
        this.sendResetCodeButtonEnable = true;
        this.resetTimer = setInterval(() => {
          this.resetCountDown--;
          this.sendResetCodeButtonName = this.resetCountDown + "秒后重新发送";
          if (this.resetCountDown === 0) {
            this.sendResetCodeButtonEnable = false;
            this.sendResetCodeButtonName = "发送重置验证码";
            this.resetCountDown = 60;
            clearInterval(this.resetTimer);
          }
        }, 1000);
      },
      
      async resetPassword() {
        this.$refs.forgotPasswordFormRef.validate(async (valid) => {
          if (!valid) return;
          
          if (this.forgotPasswordForm.newPassword !== this.forgotPasswordForm.confirmPassword) {
            return this.$message.error("两次输入的密码不一致");
          }
          
          try {
            const {data: res} = await this.$axios.post("/auth/reset-password", {
              email: this.forgotPasswordForm.email,
              emailCode: this.forgotPasswordForm.emailCode,
              newPassword: this.forgotPasswordForm.newPassword
            });
            
            if (res.status === 0) {
              this.$message.success("密码重置成功，请使用新密码登录");
              this.closeForgotPasswordDialog();
            } else {
              this.$message.error(res.message || "重置失败，请重试");
            }
          } catch (error) {
            this.$message.error("重置失败，请检查网络连接");
          }
        });
      },
      
      // OAuth相关方法
      async checkOAuthStatus() {
        try {
          const {data: res} = await this.$axios.get("/auth/oauth/status");
          if (res.status === 0) {
            this.oauthEnabled = res.data.linuxdo_enabled;
          }
        } catch (error) {
          console.log("获取OAuth状态失败:", error);
        }
      },
      
      async loginWithLinuxDo() {
        this.linuxdoLoading = true;
        try {
          const {data: res} = await this.$axios.get("/auth/oauth/linuxdo/login");
          if (res.status === 0) {
            // 跳转到LinuxDo授权页面
            window.location.href = res.data.auth_url;
          } else {
            this.$message.error(res.message || "LinuxDo登录失败");
          }
        } catch (error) {
          this.$message.error("LinuxDo登录失败，请重试");
        } finally {
          this.linuxdoLoading = false;
        }
      },
      
      // 处理OAuth回调成功
      handleOAuthSuccess() {
        const urlParams = new URLSearchParams(window.location.hash.split('?')[1]);
        const token = urlParams.get('token');
        const username = urlParams.get('username');
        const nickname = urlParams.get('nickname');
        const userId = urlParams.get('user_id');
        
        if (token && username && userId) {
          // 构建用户数据
          const userData = {
            token: token,
            username: username,
            nickname: nickname || username,
            user_id: parseInt(userId)
          };
          
          const jsonData = JSON.stringify(userData);
          const encodedData = Base64.encode(jsonData);
          
          // 保存用户信息
          window.sessionStorage.setItem("userData", encodedData);
          window.sessionStorage.setItem("token", token);
          window.sessionStorage.setItem("userId", userId); // 保存数字ID而不是username
          
          this.$message.success("LinuxDo登录成功！");
          
          // 跳转到主页
          this.$router.push("/dash");
        } else {
          this.$message.error("OAuth登录参数不完整，请重试");
        }
      }
  },
  
  mounted() {
    // 检查OAuth状态
    this.checkOAuthStatus();
    
    // 检查是否是OAuth回调成功页面
    if (this.$route.path === '/oauth/success') {
      this.handleOAuthSuccess();
    }
  }
}
</script>



<style lang="less" scoped>
.login_container {
  background: url("~@/assets/images/back7.jpg");
  width: 100%;
  height: 100%;
  background-size: 100% 100%;
  padding: 20px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;

  @media (max-width: 768px) {
    padding: 15px;
    background-size: cover;
    background-position: center;
  }
}

.login_box {
  width: 100%;
  max-width: 450px;
  min-height: 300px;
  background-color: rgba(255, 255, 255, .3);
  border-radius: 8px;
  padding: 20px;
  box-sizing: border-box;
  position: relative;

  @media (max-width: 768px) {
    max-width: 100%;
    border-radius: 12px;
    padding: 24px 20px;
    min-height: auto;
    backdrop-filter: blur(10px);
    background-color: rgba(255, 255, 255, .4);
  }

  @media (max-width: 480px) {
    padding: 20px 16px;
    margin: 0 auto;
  }
}

.login_font {
  font-size: 30px;
  font-weight: bold;
  color: #00B5AD;
  width: 100%;
  text-align: center;
  margin-bottom: 30px;

  @media (max-width: 768px) {
    font-size: 26px;
    margin-bottom: 25px;
  }

  @media (max-width: 480px) {
    font-size: 24px;
    margin-bottom: 20px;
  }
}

.login_form {
  width: 100%;

  .el-form-item {
    margin-bottom: 20px;

    @media (max-width: 768px) {
      margin-bottom: 18px;
    }
  }
}

.forgot-password-link {
  text-align: center;
  margin-top: 10px;

  .el-link {
    font-size: 14px;

    @media (max-width: 768px) {
      font-size: 15px;
      padding: 8px;
    }
  }
}

.forgot-password-form .el-form-item {
  margin-bottom: 18px;

  @media (max-width: 768px) {
    margin-bottom: 16px;
  }
}

.el-input {
  opacity: 0.8;

  @media (max-width: 768px) {
    opacity: 0.9;
  }

  :deep(.el-input__inner) {
    @media (max-width: 768px) {
      height: 44px;
      font-size: 16px;
    }
  }
}

.btns {
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;

  @media (max-width: 768px) {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
    width: 100%;
  }

  .el-button {
    min-width: 80px;

    @media (max-width: 768px) {
      height: 44px;
      font-size: 16px;
      width: 100%;
      margin: 0;
      border-radius: 8px;
      font-weight: 500;
    }
  }
}

.register_form {
  width: 100%;

  @media (max-width: 768px) {
    .el-form-item {
      margin-bottom: 16px;
    }
  }
}

.form-email {
  position: relative;

  .el-input {
    padding-right: 130px;

    @media (max-width: 768px) {
      padding-right: 0;
      margin-bottom: 10px;
    }
  }

  .el-button {
    position: absolute;
    right: 0;
    top: 0;
    height: 40px;

    @media (max-width: 768px) {
      position: static;
      width: 100%;
      height: 44px;
      margin-top: 10px;
    }
  }
}

.vcb {
  width: 120px;
  font-size: 12px;

  @media (max-width: 768px) {
    width: 100%;
    font-size: 14px;
  }
}

/* Element UI 移动端优化 */
@media (max-width: 768px) {
  :deep(.el-dialog) {
    width: 95% !important;
    margin: 0 auto !important;
    top: 5vh !important;
  }

  :deep(.el-dialog__header) {
    padding: 20px 20px 0 !important;
    
    .el-dialog__title {
      font-size: 18px !important;
    }
  }

  :deep(.el-dialog__body) {
    padding: 20px !important;
    max-height: 70vh;
    overflow-y: auto;
  }

  :deep(.el-dialog__footer) {
    padding: 10px 20px 20px !important;
    
    .el-button {
      padding: 12px 20px !important;
      min-width: 90px !important;
    }
  }

  :deep(.el-form-item__label) {
    font-size: 15px !important;
    line-height: 1.4 !important;
  }

  :deep(.el-input__inner) {
    font-size: 16px !important;
  }

  :deep(.el-message) {
    min-width: 300px !important;
    margin: 20px !important;
  }
}

/* OAuth登录样式 */
.oauth-login {
  margin-top: 20px;
  
  .oauth-divider {
    text-align: center;
    margin: 20px 0 15px;
    position: relative;
    
    &::before {
      content: '';
      position: absolute;
      top: 50%;
      left: 0;
      right: 0;
      height: 1px;
      background: rgba(255, 255, 255, 0.3);
    }
    
    span {
      background: rgba(255, 255, 255, 0.3);
      padding: 0 15px;
      color: #666;
      font-size: 14px;
      position: relative;
      z-index: 1;
    }
  }
  
  .linuxdo-login-btn {
    width: 100%;
    height: 44px;
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    color: #495057;
    font-size: 16px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s ease;
    
    &:hover {
      background: #e9ecef;
      border-color: #adb5bd;
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }
    
    &:active {
      transform: translateY(0);
    }
    
    .oauth-logo {
      width: 20px;
      height: 20px;
      margin-right: 8px;
    }
    
    @media (max-width: 768px) {
      height: 48px;
      font-size: 16px;
    }
  }
}

/* 触摸设备优化 */
@media (hover: none) and (pointer: coarse) {
  .el-button:hover {
    transform: none !important;
  }
  
  .el-button:active {
    transform: scale(0.98);
  }
  
  .linuxdo-login-btn:hover {
    transform: none !important;
    background: #e9ecef;
  }
  
  .linuxdo-login-btn:active {
    transform: scale(0.98) !important;
  }
}
</style>

