<template>
    <div class="template-view">
      <!-- Header Section -->
      <div class="page-header">
        <div class="header-content">
          <h1 class="page-title">选择开发模板</h1>
          <p class="page-subtitle">快速启动你的编程项目，支持多种主流开发环境</p>
        </div>
      </div>

      <!-- Search and Filter Bar -->
      <div class="filter-bar">
        <div class="search-section">
          <div class="search-input-wrapper">
            <i class="el-icon-search search-icon"></i>
            <input 
              v-model="searchQuery" 
              placeholder="搜索模板..." 
              class="search-input"
              @input="handleSearch"
            />
          </div>
        </div>
        
        <div class="filter-section">
          <!-- Category Filter -->
          <div class="category-filter">
            <button 
              v-for="category in allCategories" 
              :key="category.id"
              :class="['category-btn', { active: selectedCategory === category.id }]"
              @click="selectCategory(category.id)"
            >
              {{ category.name }}
            </button>
          </div>
          
          <!-- VIP Filter -->
          <div class="vip-filter">
            <button 
              :class="['filter-btn', { active: showVipOnly }]"
              @click="toggleVipFilter"
            >
              <i class="el-icon-crown"></i>
              VIP专属
            </button>
          </div>
        </div>
      </div>

      <!-- Templates Grid -->
      <div v-if="dataLoaded" class="templates-container">
        <div v-if="filteredTemplates.length === 0" class="no-results">
          <i class="el-icon-folder-opened"></i>
          <h3>没有找到匹配的模板</h3>
          <p>尝试调整搜索条件或选择其他分类</p>
        </div>
        
        <div v-else class="simple-mobile-layout" :style="{ minHeight: '400px' }">
          <!-- 调试信息 -->
          <div style="width: 100%; background: #2a2a2a; color: #fff; padding: 15px; margin-bottom: 20px; border-radius: 8px; font-size: 13px;">
            <strong>📊 模板状态:</strong> 
            共 {{ filteredTemplates.length }} 个模板可用 
            {{ filteredTemplates.length > 0 ? '✅' : '❌' }}
            <br><strong>网格信息:</strong> 屏幕宽度: {{ window.innerWidth }}px, 
            CSS Grid支持: {{ supportsGrid ? '✅' : '❌' }}, 
            isMobile: {{ isMobile }}
            <br><strong>列表:</strong> 
            <span v-for="(tmpl, index) in filteredTemplates" :key="tmpl.id">
              {{ index + 1 }}. {{ tmpl.name }}(ID:{{ tmpl.id }}){{ index < filteredTemplates.length - 1 ? ', ' : '' }}
            </span>
          </div>
          
          <TemplateCard 
            v-for="tmpl in filteredTemplates" 
            :key="tmpl.id"
            :info="tmpl" 
            :vip-info="vipInfo" 
            @select="tmplSelected"
          />
        </div>
      </div>

      <!-- Loading State -->
      <div v-else class="loading-container">
        <i class="el-icon-loading"></i>
        <p>加载模板中...</p>
      </div>


      <el-dialog custom-class="space-create-dialog" title="基本信息" :visible.sync="dialogFormVisible" width="40%" :close-on-click-modal="false" @close="onDialogClose">
        <el-form :model="spaceForm">
          <el-form-item label="空间名称:" label-width="180px">
            <el-input v-model="spaceForm.name" autocomplete="off" placeholder="请输入空间名称"></el-input>
          </el-form-item>
          <el-form-item label="空间规格:" label-width="180px">
            <!-- 移动端使用原生select，桌面端使用Element UI -->
            <el-select 
              v-if="!isMobile"
              v-model="spaceForm.space_spec_id" 
              placeholder="请选择空间规格"
              :popper-append-to-body="false"
              :popper-class="'space-spec-dropdown-fixed'"
              placement="bottom-start"
              :teleported="false"
              @visible-change="onDropdownVisibleChange">
              <el-option v-for="item in filteredSpaceSpecs" :key="item.id" :label="item.desc" :value="item.id"></el-option>
            </el-select>
            
            <!-- 移动端原生select -->
            <select 
              v-else
              v-model="spaceForm.space_spec_id" 
              class="mobile-select"
              @change="onMobileSelectChange">
              <option value="" disabled>请选择空间规格</option>
              <option v-for="item in filteredSpaceSpecs" :key="item.id" :value="item.id">{{ item.desc }}</option>
            </select>
            
            <!-- 调试信息 -->
            <div style="color: #999; font-size: 12px; margin-top: 5px;">
              调试: 可用规格数量: {{ filteredSpaceSpecs.length }}
              <span v-if="filteredSpaceSpecs.length > 0"> - {{ filteredSpaceSpecs[0].desc }}</span>
              ({{ isMobile ? '移动端' : '桌面端' }})
              <br>原始数据: {{ spaceSpecs.length }}个, 加载状态: {{ dataLoaded }}, VIP: {{ vipInfo.is_active }}
              <br>当前选择: {{ spaceForm.space_spec_id }}
            </div>
          </el-form-item>
          <!-- Git仓库配置 - Claude模板不显示 -->
          <el-form-item v-if="!selectedTemplate || selectedTemplate.id !== 7" label="Git仓库:" label-width="180px">
            <el-input v-model="spaceForm.git_repository" autocomplete="off" placeholder="请输入要克隆的Git仓库或者忽略"></el-input>
          </el-form-item>
          
          <!-- Claude模板专用配置 -->
          <div v-if="selectedTemplate && selectedTemplate.id === 7" class="claude-config">
            <el-divider content-position="left">🤖 Claude AI 配置</el-divider>
            
            <!-- API密钥 -->
            <el-form-item label="API密钥:" label-width="180px">
              <el-input 
                v-model="spaceForm.anthropic_auth_token" 
                placeholder="请输入您的API密钥 (sk-...)"
                type="password"
                show-password
                clearable
                style="width: 100%">
              </el-input>
              <div style="color: #999; font-size: 12px; margin-top: 5px;">
                <i class="el-icon-info"></i>
                输入您的Claude API密钥，支持官方和第三方服务
              </div>
            </el-form-item>
            
            <!-- API渠道选择 -->
            <el-form-item label="API渠道:" label-width="180px">
              <el-select 
                v-model="spaceForm.anthropic_base_url" 
                placeholder="选择API服务渠道"
                clearable
                style="width: 100%">
                <el-option label="🏢 Anthropic 官方 (api.anthropic.com)" value="https://api.anthropic.com"></el-option>
                <el-option label="🚀 InstCopilot 加速 (sg.instcopilot-api.com)" value="https://sg.instcopilot-api.com"></el-option>
                <el-option label="⚡ Claude Pro 代理" value="https://claude-proxy.com/v1"></el-option>
                <el-option label="🌐 Claude API Gateway" value="https://gateway.claude-api.com"></el-option>
                <el-option label="🔧 自定义地址..." value="custom"></el-option>
              </el-select>
              <div style="color: #999; font-size: 12px; margin-top: 5px;">
                <i class="el-icon-info"></i>
                选择API服务渠道，推荐使用InstCopilot加速服务
              </div>
            </el-form-item>
            
            <!-- 自定义API地址输入框 -->
            <el-form-item v-if="spaceForm.anthropic_base_url === 'custom'" label="自定义地址:" label-width="180px">
              <el-input 
                v-model="customApiUrl" 
                placeholder="请输入自定义API地址，如: https://your-api.com"
                clearable
                style="width: 100%"
                @blur="handleCustomUrlChange">
              </el-input>
              <div style="color: #999; font-size: 12px; margin-top: 5px;">
                <i class="el-icon-info"></i>
                输入完整的API基础地址，确保地址格式正确
              </div>
            </el-form-item>
          </div>
          
          <!-- VIP权限提示 -->
          <div v-if="!vipInfo.is_active" class="permission-notice">
            <p style="color: #FFA500; margin: 10px 0;">
              <i class="el-icon-warning"></i>
              普通用户只能创建测试型规格的工作空间，<span class="vip-link-disabled" style="color: #999; cursor: not-allowed; text-decoration: line-through;">升级VIP</span> 解锁更多功能（试用阶段暂未开放）
            </p>
          </div>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button type="primary" @click="createSpaceAndStart" :disabled="isCreating" :loading="isCreating">
            {{ isCreating ? '创建中...' : '创建并启动' }}
          </el-button>
          <el-button type="primary" @click="createSpace" :disabled="isCreating" :loading="isCreating">
            {{ isCreating ? '创建中...' : '创建' }}
          </el-button>
          <el-button type="info" @click="dialogFormVisible = false" :disabled="isCreating">取 消</el-button>
        </div>
      </el-dialog>
    </div>
</template>



<script>

import TemplateCard from "./TemplateCard.vue"

export default {
    components: {
        TemplateCard
    },
    data() {
        return {
          dataLoaded: false,
          spaceTemplates: [],
          spaceSpecs: [],
          dialogFormVisible: false,
          isCreating: false, // 防重复提交标志
          spaceForm: {
            name: "",
            space_spec_id: "",
            tmpl_id: 0,
            user_id: 0,
            git_repository: "",
            // Claude API 配置 - 简化版
            anthropic_auth_token: "",
            anthropic_base_url: "https://sg.instcopilot-api.com", // 默认使用InstCopilot
          },
          customApiUrl: "", // 用于自定义API地址输入
          selectedTemplate: null,
          vipInfo: {
            is_active: false,
            current_level: "普通用户",
            days_left: 0,
            expire_time: null
          },
          // 新增的搜索和过滤相关数据
          searchQuery: '',
          selectedCategory: 0, // 0表示全部分类
          showVipOnly: false,
          searchTimeout: null,
          window: window,
          supportsGrid: CSS.supports('display', 'grid')
        }
    },
    computed: {
      // 检测是否为移动端
      isMobile() {
        return window.innerWidth <= 768;
      },
      
      // 根据VIP状态过滤规格选项
      filteredSpaceSpecs() {
        console.log('=== filteredSpaceSpecs 计算 ===')
        console.log('VIP状态:', this.vipInfo.is_active)
        console.log('原始规格数据:', this.spaceSpecs)
        console.log('数据已加载:', this.dataLoaded)
        
        // 确保数据已加载
        if (!this.dataLoaded || !this.spaceSpecs || this.spaceSpecs.length === 0) {
          console.log('数据未加载完成，返回空数组')
          return [];
        }
        
        if (!this.vipInfo.is_active) {
          // 普通用户只能选择测试型规格 (ID: 4)
          const filtered = this.spaceSpecs.filter(spec => spec.id === 4);
          console.log('普通用户过滤结果:', filtered)
          return filtered;
        }
        
        console.log('VIP用户，返回所有规格:', this.spaceSpecs)
        return this.spaceSpecs;
      },
      
      // 所有分类（包含"全部"选项）
      allCategories() {
        const categories = [{ id: 0, name: '全部' }];
        this.spaceTemplates.forEach(category => {
          categories.push({ id: category.id, name: category.name });
        });
        return categories;
      },
      
      // 所有模板的扁平化列表
      allTemplates() {
        const templates = [];
        this.spaceTemplates.forEach(category => {
          category.tmpls.forEach(tmpl => {
            templates.push({
              ...tmpl,
              categoryId: category.id,
              categoryName: category.name
            });
          });
        });
        return templates;
      },
      
      // 过滤后的模板列表
      filteredTemplates() {
        let templates = this.allTemplates;
        
        // 分类过滤
        if (this.selectedCategory !== 0) {
          templates = templates.filter(tmpl => tmpl.categoryId === this.selectedCategory);
        }
        
        // VIP过滤
        if (this.showVipOnly) {
          templates = templates.filter(tmpl => tmpl.id === 7); // Claude模板
        }
        
        // 搜索过滤
        if (this.searchQuery.trim()) {
          const query = this.searchQuery.toLowerCase().trim();
          templates = templates.filter(tmpl => 
            tmpl.name.toLowerCase().includes(query) ||
            tmpl.desc.toLowerCase().includes(query) ||
            tmpl.tags.some(tag => tag.toLowerCase().includes(query))
          );
        }
        
        // 排序：Claude模板(ID=7)优先显示在第一个
        templates.sort((a, b) => {
          if (a.id === 7 && b.id !== 7) return -1; // Claude排前面
          if (a.id !== 7 && b.id === 7) return 1;  // Claude排前面
          return 0; // 其他保持原顺序
        });
        
        return templates;
      }
    },
    methods: {
        tmplSelected(id) {
          // 调试信息 - 记录选择的模板和可用规格
          console.log('=== 模板选择调试信息 ===')
          console.log('选择模板ID:', id)
          console.log('当前VIP状态:', this.vipInfo.is_active)
          console.log('原始规格数据:', this.spaceSpecs)
          console.log('过滤后的规格:', this.filteredSpaceSpecs)
          console.log('总规格数量:', this.spaceSpecs.length)
          console.log('过滤后规格数量:', this.filteredSpaceSpecs.length)
          console.log('数据是否已加载:', this.dataLoaded)
          console.log('========================')
          
          // 设置选中的模板
          this.selectedTemplate = this.allTemplates.find(t => t.id === id) || null
          
          // 清理表单数据
          this.spaceForm.tmpl_id = parseInt(id)  // 确保是数字类型
          this.spaceForm.name = ""
          this.spaceForm.space_spec_id = ""
          
          // 对于Claude模板，清空git仓库字段
          if (parseInt(id) === 7) {
            this.spaceForm.git_repository = ""
          }
          
          // 清理Claude API配置
          this.spaceForm.anthropic_auth_token = ""
          this.spaceForm.anthropic_base_url = "https://sg.instcopilot-api.com" // 重置为默认值
          this.customApiUrl = "" // 清空自定义URL
          
          // 显示对话框
          this.dialogFormVisible = true
          
          // 强制触发响应式更新
          this.$forceUpdate()
          
          // 在下一个tick中强制刷新下拉框
          this.$nextTick(() => {
            console.log('对话框已显示，检查规格选项...')
            console.log('选择框中的规格:', this.filteredSpaceSpecs)
            console.log('强制重新计算过滤规格...')
            
            // 强制重新计算filteredSpaceSpecs
            const specs = this.filteredSpaceSpecs
            console.log('重新计算后的规格:', specs)
            
            // 清理可能存在的重复下拉框
            const existingDropdowns = document.querySelectorAll('.el-select-dropdown')
            console.log('发现下拉框数量:', existingDropdowns.length)
            if (existingDropdowns.length > 1) {
              console.log('清理多余的下拉框...')
              for (let i = 1; i < existingDropdowns.length; i++) {
                existingDropdowns[i].remove()
              }
            }
          })
        },
        
        // 加载VIP信息
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
        
        // 跳转到VIP升级页面
        goToVip() {
          // 试用阶段：禁用VIP功能
          this.$message.info('试用阶段暂未开放VIP功能')
        },
        
        // 对话框关闭时清理残留的下拉框
        onDialogClose() {
          this.$nextTick(() => {
            const existingDropdowns = document.querySelectorAll('.el-select-dropdown')
            console.log('对话框关闭，清理下拉框，数量:', existingDropdowns.length)
            existingDropdowns.forEach(dropdown => {
              if (dropdown.style.display !== 'none') {
                dropdown.remove()
              }
            })
          })
        },
        
        // 下拉框可见性变化监听
        onDropdownVisibleChange(visible) {
          console.log('=== 下拉框可见性变化 ===')
          console.log('下拉框可见:', visible)
          console.log('当前选择的模板ID:', this.spaceForm.tmpl_id)
          console.log('VIP状态:', this.vipInfo.is_active)
          console.log('可用规格:', this.filteredSpaceSpecs)
          console.log('原始规格数据:', this.spaceSpecs)
          console.log('数据加载状态:', this.dataLoaded)
          console.log('=======================')
          
          // 修复下拉框定位问题
          if (visible) {
            this.$nextTick(() => {
              // 查找下拉框元素并修复定位
              const dropdown = document.querySelector('.space-spec-dropdown-fixed')
              if (dropdown) {
                console.log('找到下拉框，修复定位...')
                // 重置transform和position
                dropdown.style.transform = 'none'
                dropdown.style.willChange = 'auto'
                dropdown.style.position = 'absolute'
                
                // 确保z-index足够高
                dropdown.style.zIndex = '2060'
                
                // 如果还是有问题，尝试重新计算位置
                const selectElement = dropdown.parentElement?.querySelector('.el-select')
                if (selectElement) {
                  const rect = selectElement.getBoundingClientRect()
                  console.log('Select元素位置:', rect)
                }
              }
            })
          }
        },
        
        // 移动端原生select变化监听
        onMobileSelectChange(event) {
          console.log('=== 移动端选择变化 ===')
          console.log('选择的值:', event.target.value)
          console.log('值的类型:', typeof event.target.value)
          console.log('当前模板ID:', this.spaceForm.tmpl_id)
          // 确保转为数字类型
          this.spaceForm.space_spec_id = parseInt(event.target.value)
          console.log('转换后的值:', this.spaceForm.space_spec_id)
          console.log('转换后的类型:', typeof this.spaceForm.space_spec_id)
          console.log('====================')
        },
        
        // 窗口大小变化处理
        handleResize() {
          this.$forceUpdate()
        },
        
        // 处理搜索输入（防抖）
        handleSearch() {
          if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
          }
          this.searchTimeout = setTimeout(() => {
            // 搜索逻辑在computed属性中处理
          }, 300);
        },
        
        // 选择分类
        selectCategory(categoryId) {
          this.selectedCategory = categoryId;
        },
        
        // 切换VIP过滤
        toggleVipFilter() {
          this.showVipOnly = !this.showVipOnly;
        },
        
        // 重置过滤器
        resetFilters() {
          this.searchQuery = '';
          this.selectedCategory = 0;
          this.showVipOnly = false;
        },
        
        joinPath(p1, p2) {
            return p1.replace(/\/$/, '') + "/" + p2.replace(/^\//, '');
        },
        
        async getTemplates() {
            const {data: res} = await this.$axios.get("/api/template/list")
            if (res.status) {
                this.$message.error(res.message)
                return
            }
            const kinds = res.data.kinds
            const tmpls = res.data.tmpls.sort((a, b) => {
                return a.id - b.id
            })
            
            // 重置模板数组
            this.spaceTemplates = []
            
            kinds.forEach((ele, index) => {
                // 为每个分类创建新的对象
                this.spaceTemplates[index] = {
                    id: ele.id,
                    name: ele.name,
                    tmpls: []
                }
                
                for (let i = 0; i < tmpls.length; i++) {
                    if (ele.id === tmpls[i].kind_id) {
                        var t = tmpls[i]
                        const tags = t.tags.split(',')
                        this.spaceTemplates[index].tmpls.push({...t, tags})
                        for (let j = 0; j < this.spaceTemplates[index].tmpls.length; j++) {
                            const avatar = this.spaceTemplates[index].tmpls[j].avatar
                            if (!avatar.startsWith("http") && !avatar.startsWith("https")) {
                                // 使用当前域名构建完整URL
                                const currentDomain = window.location.origin
                                this.spaceTemplates[index].tmpls[j].avatar = `${currentDomain}/${avatar}`
                            }
                        }
                    }
                }
            })
            
            console.log('模板数据处理完成:', this.spaceTemplates)
            
            // 专门调试所有模板的URL
            this.spaceTemplates.forEach(category => {
                category.tmpls.forEach(tmpl => {
                    console.log(`📝 模板 ${tmpl.name} (ID:${tmpl.id}):`)
                    console.log(`  - 完整URL: ${tmpl.avatar}`)
                    if (tmpl.id === 7) {
                        console.log('🔍 ⬆️ 这是Claude模板')
                    }
                })
            })
      },
      
      // 验证权限
      validatePermission() {
        // 试用阶段：Claude模板开放给所有用户，移除VIP权限检查
        // if (this.spaceForm.tmpl_id === 7 && !this.vipInfo.is_active) {
        //   this.$message.warning("Claude AI助手功能仅限VIP用户使用，请升级为VIP用户");
        //   return false;
        // }
        
        // 检查规格权限
        if (this.spaceForm.space_spec_id !== 4 && !this.vipInfo.is_active) {
          this.$message.warning("普通用户只能创建测试型配置的工作空间，请升级为VIP用户使用其他配置");
          return false;
        }
        
        return true;
      },
      
      validateCreateInfo() {
        if (!(this.spaceForm.name.trim())) {
          this.$message.warning("请输入要创建的工作空间的名称")
          return false
        }
        const value = this.spaceForm.name
        const chineseMatch = value.match(/[\u4e00-\u9fa5]/g)
        const englishMatch = value.match(/[a-zA-Z]/g)
        let chineseCount = 0
        let englishCount = 0
        if (chineseMatch) {
            chineseCount = chineseMatch.length
        }
        if (englishMatch) {
            englishCount = englishMatch.length
        }
        if (chineseCount * 2 + englishCount > 32) {
            this.$message.warning("名称的长度过长,中文字符最多16个,英文字符最多32个")
            return false
        }

        if (!this.spaceForm.space_spec_id) {
          this.$message.warning("请选择要创建的工作空间的规格")
          return false
        }
        
        // 权限验证
        if (!this.validatePermission()) {
          return false
        }
        
        console.log("验证git仓库：", this.spaceForm.git_repository)
        const regex = /^https:\/\/\S+\.git$/
        this.spaceForm.git_repository = this.spaceForm.git_repository.trim()
        if (this.spaceForm.git_repository.length === 0) {
          return true
        }
        console.log("验证git仓库：", this.spaceForm.git_repository)
        if (!regex.test(this.spaceForm.git_repository.trim())) {
          console.log("git地址无效")
          this.$message.warning("请输入有效的Git仓库地址")
          this.spaceForm.git_repository = ""
          return false
        }
        
        console.log("git正则通过")
        return true
      },
      async getSpaceSpecs() {
        console.log('开始加载空间规格数据...')
        try {
        const {data:res} = await this.$axios.get("/api/spec/list")
        if (res.status) {
          this.$message.error(res.message)
            console.log('空间规格加载失败:', res.message)
          return
        }
          console.log('空间规格API返回数据:', res.data)
          this.spaceSpecs = res.data || []
          console.log('空间规格设置完成:', this.spaceSpecs)
        } catch (error) {
          console.error('加载空间规格时出错:', error)
          // 出错时使用默认数据
          this.spaceSpecs = [
            {id: 4, cpu_spec: 2, mem_spec: '2Gi', storage_spec: '4Gi', name: '测试型', desc: '测试型 2CPU 2GB内存 / 4GB存储'}
          ]
        }
      },
      async createSpaceAndStart() {
        // 防重复提交检查
        if (this.isCreating) {
          this.$message.warning('工作空间正在创建中，请稍候...')
          return
        }

        // 先进行权限验证，避免无权限时也显示loading
        if (!this.validateCreateInfo()) {
          return          
        }

        // 设置创建中状态
        this.isCreating = true

        // 添加详细的参数调试
        console.log('=== 准备创建并启动空间 ===')
        console.log('提交的表单数据:', JSON.stringify(this.spaceForm, null, 2))
        console.log('各字段类型检查:')
        console.log('  name:', this.spaceForm.name, '(type:', typeof this.spaceForm.name, ')')
        console.log('  space_spec_id:', this.spaceForm.space_spec_id, '(type:', typeof this.spaceForm.space_spec_id, ')')
        console.log('  tmpl_id:', this.spaceForm.tmpl_id, '(type:', typeof this.spaceForm.tmpl_id, ')')
        console.log('  user_id:', this.spaceForm.user_id, '(type:', typeof this.spaceForm.user_id, ')')
        console.log('  git_repository:', this.spaceForm.git_repository, '(type:', typeof this.spaceForm.git_repository, ')')
        console.log('============================')

        this.dialogFormVisible = false

        const loading = this.$loading({
            lock: true,
            text: '正在创建工作空间，请勿重复操作...',
            spinner: 'el-icon-loading',
            background: 'rgba(0, 0, 0, 0.7)'
        });

        try {
        const {data:res} = await this.$axios.post("/api/workspace/cas", this.spaceForm)
        if (res.status) {
            console.log('创建空间失败:', res.message)
            console.log('完整响应:', JSON.stringify(res, null, 2))
          this.$message.error(res.message)
          loading.close()
          this.isCreating = false
          return
        }

        setTimeout(() => {
          loading.close()
          this.isCreating = false
          const spaceUrl =  this.$axios.defaults.workspaceUrl + res.data.sid + "/"
          window.open(spaceUrl, '_blank')
        }, 2000);
        } catch (error) {
          console.error('API调用出错:', error)
          console.log('错误详情:', error.response?.data)
          this.$message.error('创建空间时发生错误: ' + (error.response?.data?.message || error.message))
          loading.close()
          this.isCreating = false
        }
        
      },
      async createSpace() {
        // 防重复提交检查
        if (this.isCreating) {
          this.$message.warning('工作空间正在创建中，请稍候...')
          return
        }

        // 先进行权限验证，避免无权限时也显示loading
        if (!this.validateCreateInfo()) {
          return          
        }

        // 设置创建中状态
        this.isCreating = true

        // 添加详细的参数调试
        console.log('=== 准备创建空间 ===')
        console.log('提交的表单数据:', JSON.stringify(this.spaceForm, null, 2))
        console.log('各字段类型检查:')
        console.log('  name:', this.spaceForm.name, '(type:', typeof this.spaceForm.name, ')')
        console.log('  space_spec_id:', this.spaceForm.space_spec_id, '(type:', typeof this.spaceForm.space_spec_id, ')')
        console.log('  tmpl_id:', this.spaceForm.tmpl_id, '(type:', typeof this.spaceForm.tmpl_id, ')')
        console.log('  user_id:', this.spaceForm.user_id, '(type:', typeof this.spaceForm.user_id, ')')
        console.log('  git_repository:', this.spaceForm.git_repository, '(type:', typeof this.spaceForm.git_repository, ')')
        console.log('=====================')

        this.dialogFormVisible = false

        try {
        const {data:res} = await this.$axios.post("/api/workspace", this.spaceForm)
        if (res.status) {
            console.log('创建空间失败:', res.message)
            console.log('完整响应:', JSON.stringify(res, null, 2))
          this.$message.error(res.message)
        } else {
            console.log('创建空间成功:', res.message)
          this.$message.success(res.message)
          }
        } catch (error) {
          console.error('API调用出错:', error)
          console.log('错误详情:', error.response?.data)
          this.$message.error('创建空间时发生错误: ' + (error.response?.data?.message || error.message))
        } finally {
          this.isCreating = false
        }
      },
      
      // 处理自定义API地址变化
      handleCustomUrlChange() {
        if (this.customApiUrl && this.customApiUrl.trim()) {
          // 验证URL格式
          try {
            new URL(this.customApiUrl.trim())
            this.spaceForm.anthropic_base_url = this.customApiUrl.trim()
          } catch (e) {
            this.$message.error('请输入有效的URL地址')
            this.customApiUrl = ""
          }
        }
      }
      
    },
    
    async mounted() {
      console.log('=== 组件挂载开始 ===')
      this.spaceForm.user_id = parseInt(window.sessionStorage.getItem("userId"))
      console.log('用户ID:', this.spaceForm.user_id)
      
      console.log('开始加载VIP信息...')
      await this.loadVipInfo()
      console.log('VIP信息加载完成:', this.vipInfo)
      
      console.log('开始加载模板...')
      await this.getTemplates()
      console.log('模板加载完成')
      console.log('🔍 当前模板数据结构:', this.spaceTemplates)
      
      console.log('开始加载空间规格...')
      await this.getSpaceSpecs()
      console.log('空间规格加载完成:', this.spaceSpecs)
      
      this.dataLoaded = true
      console.log('所有数据加载完成，dataLoaded设为true')
      
      // 添加窗口大小变化监听器
      window.addEventListener('resize', this.handleResize)
      
      // 最终状态检查
      console.log('=== 最终状态检查 ===')
      console.log('VIP状态:', this.vipInfo.is_active)
      console.log('原始规格数据:', this.spaceSpecs)
      console.log('过滤后的规格:', this.filteredSpaceSpecs)
      console.log('数据加载标志:', this.dataLoaded)
      console.log('是否移动端:', this.isMobile)
      console.log('==================')
    },
    
    beforeDestroy() {
      window.removeEventListener('resize', this.handleResize)
    }
}
</script>



<style lang="scss" scoped>
@import '../assets/mobile-responsive.css';

.template-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1d21 0%, #2c3034 100%);
  padding: 0;
  position: relative;
  overflow-x: hidden; /* 防止水平滚动 */
}

// Page Header
.page-header {
  background: linear-gradient(135deg, #2c3034 0%, #1a1d21 100%);
  padding: 40px 24px;
  text-align: center;
  border-bottom: 1px solid #3a3f47;
}

.header-content {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 12px 0;
  background: linear-gradient(135deg, #409eff 0%, #67c23a 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.page-subtitle {
  font-size: 1.1rem;
  color: #b8bcc5;
  margin: 0;
  line-height: 1.6;
}

// Filter Bar
.filter-bar {
  background: rgba(42, 46, 52, 0.8);
  backdrop-filter: blur(10px);
  padding: 20px 24px;
  border-bottom: 1px solid #3a3f47;
  position: sticky;
  top: 0;
  z-index: 10;
}

.search-section {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.search-input-wrapper {
  position: relative;
  max-width: 400px;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: #9aa0a9;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 12px 16px 12px 48px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  color: #ffffff;
  font-size: 14px;
  transition: all 0.3s ease;

  &::placeholder {
    color: #9aa0a9;
  }

  &:focus {
    outline: none;
    border-color: #409eff;
    background: rgba(64, 158, 255, 0.1);
    box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
  }
}

.filter-section {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
}

.category-filter {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.category-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;

  &:hover {
    background: rgba(64, 158, 255, 0.2);
    border-color: #409eff;
  }

  &.active {
    background: linear-gradient(135deg, #409eff 0%, #67c23a 100%);
    border-color: transparent;
    color: white;
  }
}

.vip-filter {
  display: flex;
  align-items: center;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.3);
  border-radius: 20px;
  color: #ffd700;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;

  &:hover {
    background: rgba(255, 215, 0, 0.2);
    border-color: #ffd700;
  }

  &.active {
    background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
    border-color: transparent;
    color: #2c2c2c;
  }

  i {
    font-size: 16px;
  }
}

// Templates Container
.templates-container {
  padding: 32px 24px;
  max-width: 1400px;
  margin: 0 auto;
  position: relative;
  width: 100%;
  box-sizing: border-box;
}

/* 保留原始 templates-grid 类名但不使用 */
.templates-grid {
  /* 已迁移到 .simple-templates-layout */
}

// No Results State
.no-results {
  text-align: center;
  padding: 60px 20px;
  color: #9aa0a9;

  i {
    font-size: 4rem;
    margin-bottom: 20px;
    opacity: 0.5;
  }

  h3 {
    font-size: 1.5rem;
    color: #ffffff;
    margin: 0 0 12px 0;
  }

  p {
    font-size: 1rem;
    margin: 0;
    line-height: 1.6;
  }
}

// Loading State
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #9aa0a9;

  i {
    font-size: 2rem;
    margin-bottom: 16px;
    animation: spin 1s linear infinite;
  }

  p {
    font-size: 1rem;
    margin: 0;
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// Mobile Optimizations
@media (max-width: 768px) {
  .page-header {
    padding: 24px 16px;
  }

  .page-title {
    font-size: 2rem;
  }

  .page-subtitle {
    font-size: 1rem;
  }

  .filter-bar {
    padding: 16px;
  }

  .search-section {
    margin-bottom: 16px;
  }

  .filter-section {
    gap: 16px;
  }

  .category-filter {
    gap: 6px;
  }

  .category-btn {
    padding: 6px 12px;
    font-size: 13px;
  }

  .filter-btn {
    padding: 6px 12px;
    font-size: 13px;
  }

  .templates-container {
    padding: 20px 16px;
  }
}

@media (max-width: 480px) {
  .page-header {
    padding: 20px 12px;
  }

  .page-title {
    font-size: 1.75rem;
  }

  .filter-bar {
    padding: 12px;
  }

  .filter-section {
    flex-direction: column;
    gap: 12px;
}

  .templates-container {
    padding: 16px 12px;
  }
}
</style>

<!-- Dialog Styles (Global) -->
<style lang="scss">
.space-create-dialog {
  background-color: #323640 !important;
  
  .el-form-item__label {
    color: #FFF;
  }

  .el-input__inner {
    background-color: #3C414C;
    border-color: #494D57;
    color: #cfcdcd;
  }
  
  .permission-notice {
    background-color: #2B2B2B;
    padding: 10px;
    border-radius: 4px;
    border-left: 4px solid #FFA500;
    margin: 10px 0;
  }
}

.el-scrollbar__view, .el-select-dropdown__item {
  background-color: #3C414D !important;
  border-color: #494D57;
  color: #dfdede !important;
}

.el-select-dropdown {
  border: none !important;
}

.popper__arrow::after {
  border-bottom-color: #3C414D !important;
}

.el-select-dropdown__item:hover {
  background: #6e6180 !important;
}

.el-dialog {
  .el-form {
    width: 76%;
  }

  .el-input {
    width: 100%;
  }

  .el-select {
    width: 100%;
  }
}

/* 对话框移动端优化 */
@media (max-width: 768px) {
  .space-create-dialog {
    width: 95% !important;
    margin: 0 auto !important;

    .el-dialog__header {
      padding: 20px 20px 0 !important;
    }

    .el-dialog__body {
      padding: 20px !important;
    }

    .el-form {
      width: 100% !important;

      .el-form-item {
        margin-bottom: 20px;

        &__label {
          font-size: 15px !important;
          line-height: 1.4 !important;
          margin-bottom: 8px;
        }
      }
    }

    .el-input__inner,
    .el-select .el-input__inner {
      height: 44px !important;
      font-size: 16px !important;
}

    .dialog-footer {
      text-align: center;

      .el-button {
        padding: 12px 20px !important;
        margin: 5px !important;
        min-width: 90px;

        @media (max-width: 480px) {
          width: 100%;
          margin: 8px 0 !important;
        }
      }
    }
}

  .permission-notice {
    padding: 15px !important;
    margin: 15px 0 !important;

    p {
      font-size: 14px !important;
      line-height: 1.5 !important;
    }
  }
}

/* 简化的下拉框修复 - 避免复杂的定位问题 */
.el-select-dropdown {
  z-index: 2050 !important;
}

.el-popper {
  z-index: 2050 !important;
}

/* 空间规格下拉框特殊样式 */
.space-spec-dropdown {
  background-color: #3C414D !important;
  border: 1px solid #494D57 !important;
  border-radius: 4px !important;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.4) !important;
  
  .el-select-dropdown__item {
    background-color: #3C414D !important;
    color: #dfdede !important;
    padding: 8px 20px !important;
    
    &:hover {
      background-color: #6e6180 !important;
    }
    
    &.selected {
      background-color: #409EFF !important;
      color: #fff !important;
    }
  }
}

/* 修复滑动后下拉框失位问题 */
.space-spec-dropdown-fixed {
  position: absolute !important;
  transform: none !important;
  will-change: auto !important;
  
  /* 确保下拉框在正确的位置 */
  &.el-select-dropdown {
    position: absolute !important;
    transform: translateY(0) !important;
    margin-top: 5px !important;
  }
  
  /* 继承原有样式 */
  background-color: #3C414D !important;
  border: 1px solid #494D57 !important;
  border-radius: 4px !important;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.4) !important;
  z-index: 2060 !important;
  
  .el-select-dropdown__item {
    background-color: #3C414D !important;
    color: #dfdede !important;
    padding: 8px 20px !important;
    
    &:hover {
      background-color: #6e6180 !important;
    }
    
    &.selected {
      background-color: #409EFF !important;
      color: #fff !important;
    }
  }
}

/* 对话框内的定位修复 */
.el-dialog {
  .el-select-dropdown {
    position: absolute !important;
    transform: none !important;
  }
}

/* 模板视图容器的定位修复 */
.template-view {
  position: relative !important;
  
  .el-form {
    position: relative !important;
    z-index: 1 !important;
  }
  
  .el-select {
    position: relative !important;
    z-index: 10 !important;
  }
}

/* 简单移动端风格布局 - 完全模拟移动端行为 */
.simple-mobile-layout {
  /* 始终使用flexbox，避免CSS Grid的复杂性 */
  display: flex !important;
  flex-direction: column !important;
  align-items: stretch !important;
  gap: 0 !important;
  
  /* 调试信息全宽显示 */
  > div:first-child {
    width: 100% !important;
    margin-bottom: 20px !important;
  }
  
  /* 模板卡片容器 */
  .template-card {
    width: 100% !important;
    max-width: none !important;
    margin: 0 0 20px 0 !important;
    flex: none !important;
  }
}

/* 桌面端优化 - 使用简单的多列布局 */
@media (min-width: 769px) {
  .simple-mobile-layout {
    flex-direction: row !important;
    flex-wrap: wrap !important;
    justify-content: flex-start !important;
    align-items: flex-start !important;
    
    > div:first-child {
      flex: 0 0 100% !important;
      width: 100% !important;
    }
    
    .template-card {
      flex: 0 0 calc(33.333% - 14px) !important;
      width: calc(33.333% - 14px) !important;
      max-width: calc(33.333% - 14px) !important;
      margin: 0 7px 20px 7px !important;
    }
  }
}

@media (min-width: 1200px) {
  .simple-mobile-layout .template-card {
    flex: 0 0 calc(25% - 15px) !important;
    width: calc(25% - 15px) !important;
    max-width: calc(25% - 15px) !important;
    margin: 0 7.5px 20px 7.5px !important;
  }
}

/* 移动端原生select样式 */
.mobile-select {
  width: 100%;
  height: 44px;
  padding: 8px 12px;
  font-size: 16px;
  color: #cfcdcd;
  background-color: #3C414C;
  border: 1px solid #494D57;
  border-radius: 4px;
  outline: none;
  
  option {
    background-color: #3C414C;
    color: #cfcdcd;
    padding: 8px;
  }
  
  &:focus {
    border-color: #409EFF;
  }
}

/* 移动端特殊处理 */
@media (max-width: 768px) {
  .el-select-dropdown__item {
    padding: 12px 20px !important;
    font-size: 16px !important;
    line-height: 1.4 !important;
    background-color: #3C414D !important;
    color: #dfdede !important;
    
    &:hover {
      background: #6e6180 !important;
    }
  }
}
</style>