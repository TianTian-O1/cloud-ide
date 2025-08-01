
<template>
    <div 
        class="space-card" 
        :class="{ 
            'running': space.running_status,
            'swipe-left': swipeDirection === 'left',
            'long-pressing': longPressing
        }"
        @touchstart="handleTouchStart"
        @touchmove="handleTouchMove"
        @touchend="handleTouchEnd"
        @click="handleCardClick"
    >
        <!-- 滑动删除背景 -->
        <div class="swipe-background" v-if="swipeDirection === 'left'">
            <div class="swipe-action">
                <i class="el-icon-delete"></i>
                <span>删除</span>
            </div>
        </div>
        
        <div class="card-content" ref="cardContent">
            <!-- Logo区域 -->
            <div class="logo-section">
                <div class="logo-container">
                    <img class="workspace-logo" :src="space.avatar" alt="工作空间图标">
                    <div class="status-indicator" v-if="space.running_status" :class="{ 'active': space.running_status }">
                        <div class="status-dot"></div>
                    </div>
                </div>
            </div>
            
            <!-- 信息区域 -->
            <div class="info-section">
                <div class="workspace-info">
                    <div class="primary-info">
                        <h3 class="workspace-name">{{ space.name }}</h3>
                        <div class="workspace-status">
                            <span class="status-text" :class="{ 'running': space.running_status }">
                                {{ space.running_status ? '运行中' : '已停止' }}
                            </span>
                        </div>
                    </div>
                    
                    <div class="secondary-info">
                        <div class="info-item environment">
                            <i class="el-icon-cpu"></i>
                            <span class="value">{{ space.environment }}</span>
                        </div>
                        <div class="info-item spec">
                            <i class="el-icon-monitor"></i>
                            <span class="value">{{ spaceSpecDesc }}</span>
                        </div>
                        <div class="info-item time">
                            <i class="el-icon-time"></i>
                            <span class="value">{{ space.create_time | dateFormat }}</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- 操作区域 -->
            <div class="actions-section">
                <div class="action-buttons">
                    <!-- 主要操作按钮 -->
                    <div class="primary-actions">
                        <el-button 
                            v-if="space.running_status"
                            type="success" 
                            size="small" 
                            class="primary-btn enter-btn"
                            @click.stop="enterWorkspace"
                        >
                            <i class="el-icon-right"></i>
                            <span class="btn-text">进入</span>
                        </el-button>
                        <el-button 
                            v-else
                            type="primary" 
                            size="small" 
                            class="primary-btn start-btn"
                            @click.stop="startWorkspace"
                        >
                            <i class="el-icon-video-play"></i>
                            <span class="btn-text">启动</span>
                        </el-button>
                    </div>
                    
                    <!-- 次要操作按钮 -->
                    <div class="secondary-actions">
                        <el-tooltip effect="dark" content="停止" placement="top" v-if="space.running_status">
                            <el-button 
                                size="mini" 
                                circle 
                                class="icon-btn stop-btn"
                                @click.stop="stopWorkspace"
                            >
                                <i class="el-icon-video-pause"></i>
                            </el-button>
                        </el-tooltip>
                        
                        <el-tooltip effect="dark" content="编辑名称" placement="top" v-if="!space.running_status">
                            <el-button 
                                size="mini" 
                                circle 
                                class="icon-btn edit-btn"
                                @click.stop="openEditDialog"
                            >
                                <i class="el-icon-edit"></i>
                            </el-button>
                        </el-tooltip>
                        
                        <el-tooltip effect="dark" content="删除" placement="top">
                            <el-button 
                                size="mini" 
                                circle 
                                type="danger"
                                class="icon-btn delete-btn"
                                @click.stop="deleteWorkspace"
                            >
                                <i class="el-icon-delete"></i>
                            </el-button>
                        </el-tooltip>
                    </div>
                </div>
            </div>
        </div>

        <!-- 长按菜单 -->
        <div 
            v-if="showContextMenu && isMobile" 
            class="context-menu"
            :style="{ top: contextMenuY + 'px', left: contextMenuX + 'px' }"
        >
            <div class="menu-item" @click="enterOrStartWorkspace">
                <i :class="space.running_status ? 'el-icon-right' : 'el-icon-video-play'"></i>
                {{ space.running_status ? '进入工作空间' : '启动工作空间' }}
            </div>
            <div v-if="space.running_status" class="menu-item" @click="stopWorkspace">
                <i class="el-icon-video-pause"></i>
                停止工作空间
            </div>
            <div v-if="!space.running_status" class="menu-item" @click="openEditDialog">
                <i class="el-icon-edit"></i>
                编辑名称
            </div>
            <div class="menu-item danger" @click="deleteWorkspace">
                <i class="el-icon-delete"></i>
                删除工作空间
            </div>
        </div>

        <!-- 遮罩层用于关闭菜单 -->
        <div 
            v-if="showContextMenu" 
            class="context-menu-overlay" 
            @click="closeContextMenu"
            @touchstart="closeContextMenu"
        ></div>
    </div>
</template>

<script>
import "../assets/icon/iconfont.css"

export default {
    props: ["space", "index"],
    data() {
        return {
            // 触摸相关
            touchStartX: 0,
            touchStartY: 0,
            touchStartTime: 0,
            touchMoved: false,
            swipeDirection: null,
            swipeDistance: 0,
            longPressing: false,
            longPressTimer: null,
            
            // 菜单相关
            showContextMenu: false,
            contextMenuX: 0,
            contextMenuY: 0,
            
            // 设备检测
            isMobile: false
        }
    },
    computed: {
        spaceSpecDesc() {
            const spec = this.space.spec
            const strs = spec.mem_spec.split('i')
            const desc = spec.name + " " + spec.cpu_spec + "C" + strs[0]
            return desc
        }
    },
    mounted() {
        this.checkMobile()
        window.addEventListener('resize', this.checkMobile)
        document.addEventListener('click', this.handleDocumentClick)
    },
    beforeDestroy() {
        window.removeEventListener('resize', this.checkMobile)
        document.removeEventListener('click', this.handleDocumentClick)
        if (this.longPressTimer) {
            clearTimeout(this.longPressTimer)
        }
    },
    methods: {
        checkMobile() {
            this.isMobile = window.innerWidth <= 768
        },
        
        // 触摸事件处理
        handleTouchStart(e) {
            if (!this.isMobile) return
            
            const touch = e.touches[0]
            this.touchStartX = touch.clientX
            this.touchStartY = touch.clientY
            this.touchStartTime = Date.now()
            this.touchMoved = false
            this.swipeDirection = null
            this.swipeDistance = 0
            
            // 长按检测
            this.longPressTimer = setTimeout(() => {
                if (!this.touchMoved) {
                    this.handleLongPress(e)
                }
            }, 600) // 600ms 长按
        },
        
        handleTouchMove(e) {
            if (!this.isMobile) return
            
            const touch = e.touches[0]
            const deltaX = touch.clientX - this.touchStartX
            const deltaY = touch.clientY - this.touchStartY
            
            // 计算移动距离和方向
            const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY)
            
            if (distance > 10) {
                this.touchMoved = true
                if (this.longPressTimer) {
                    clearTimeout(this.longPressTimer)
                    this.longPressing = false
                }
                
                // 水平滑动检测
                if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 30) {
                    this.swipeDistance = Math.abs(deltaX)
                    
                    if (deltaX < -50) { // 向左滑动
                        this.swipeDirection = 'left'
                        e.preventDefault() // 阻止默认滚动行为
                    } else {
                        this.swipeDirection = null
                    }
                }
            }
        },
        
        handleTouchEnd(e) {
            if (!this.isMobile) return
            
            if (this.longPressTimer) {
                clearTimeout(this.longPressTimer)
            }
            
            this.longPressing = false
            
            // 处理滑动删除
            if (this.swipeDirection === 'left' && this.swipeDistance > 100) {
                this.handleSwipeDelete()
            } else {
                // 重置滑动状态
                setTimeout(() => {
                    this.swipeDirection = null
                    this.swipeDistance = 0
                }, 150)
            }
        },
        
        handleLongPress(e) {
            if (!this.isMobile) return
            
            this.longPressing = true
            
            // 获取触摸位置
            const touch = e.touches[0]
            this.contextMenuX = Math.min(touch.clientX, window.innerWidth - 200)
            this.contextMenuY = Math.min(touch.clientY, window.innerHeight - 200)
            
            this.showContextMenu = true
            
            // 振动反馈（如果支持）
            if (navigator.vibrate) {
                navigator.vibrate(50)
            }
        },
        
        handleCardClick(e) {
            if (this.touchMoved || this.showContextMenu) {
                e.preventDefault()
                return
            }
            
            // 默认点击行为：运行中的进入，停止的启动
            if (this.space.running_status) {
                this.enterWorkspace()
            } else {
                this.startWorkspace()
            }
        },
        
        handleSwipeDelete() {
            this.$messageBox.confirm(
                '确定要删除这个工作空间吗？',
                '滑动删除',
                {
                    confirmButtonText: '删除',
                    cancelButtonText: '取消',
                    type: 'warning',
                    customClass: 'mobile-confirm-dialog'
                }
            ).then(() => {
                this.deleteWorkspace()
            }).catch(() => {
                // 重置滑动状态
                this.swipeDirection = null
                this.swipeDistance = 0
            })
        },
        
        closeContextMenu() {
            this.showContextMenu = false
        },
        
        handleDocumentClick(e) {
            if (this.showContextMenu && !this.$el.contains(e.target)) {
                this.closeContextMenu()
            }
        },
        
        enterOrStartWorkspace() {
            this.closeContextMenu()
            if (this.space.running_status) {
                this.enterWorkspace()
            } else {
                this.startWorkspace()
            }
        },

        // 原有方法保持不变
        enterWorkspace() {
            if (this.space.running_status) {
                const url = this.$axios.defaults.workspaceUrl + this.space.sid + "/"
                window.open(url, "_blank")
            }
        },
        async startWorkspace() {
            if (this.space.running_status) {
                return
            }
            // 记载中动画
            const loading = this.$loading({
                lock: true,
                text: '正在启动工作空间...',
                spinner: 'el-icon-loading',
                background: 'rgba(0, 0, 0, 0.7)'
            });

            try {
                const {data:res} = await this.$axios.put("/api/workspace/start", {id: this.space.id})
                if (res.status) {
                    this.$message.error(res.message)
                    loading.close()
                    return
                }

                // 2s钟后在打开
                setTimeout(() => {
                    loading.close()
                    this.$message.success(res.message)
                    const url = this.$axios.defaults.workspaceUrl + res.data.sid + "/"
                    window.open(url, "_blank")
                    // 通知父组件改变space的running_status字段
                    this.$emit("onStartSpace", this.index, true)
                }, 2000);           
            } catch (error) {
                loading.close()
                this.$message.error('启动失败')
            }
        },
        async stopWorkspace() {
            this.closeContextMenu()
            
            if (!this.space.running_status) {
                return
            }
            
            try {
                const {data:res} = await this.$axios.put("/api/workspace/stop", {id: this.space.id})
                if (res.status) {
                    this.$message.error(res.message)
                    return
                }

                this.$message.success(res.message)
                this.$emit("onStopSpace", this.index, false)
            } catch (error) {
                this.$message.error('停止失败')
            }
        },
        deleteWorkspace(){
            this.closeContextMenu()
            
            this.$messageBox({
                title: "删除确认",
                message: "确定删除此工作空间吗？删除后无法恢复！",
                type: "warning",
                showCancelButton: true,
                confirmButtonText: '确定删除',
                cancelButtonText: '取消',
                customClass: "delete-confirm"
            }).then(async () => {
                
                if (this.space.running_status) {
                    this.$message.warning("工作空间正在运行,请先停止!")
                    return
                }
                
                try {
                    const {data:res} = await this.$axios.delete("/api/workspace", {data: {id: this.space.id}})
                    if (res.status) {
                        this.$message.error(res.message)
                        return
                    }
                    
                    this.$message.success(res.message)
                    this.$emit("onDeleteSpace", this.index)
                } catch (error) {
                    this.$message.error('删除失败')
                }
            }, () => {
                this.$message({type: 'info', message: '已取消删除'});
            });
        },
        openEditDialog() {
            this.closeContextMenu()
            
            this.$messageBox.prompt('请输入新的工作空间名称', '修改工作空间名称', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    customClass: "delete-confirm",
                    inputValue: this.space.name
                }).then(({ value }) => {
                    if (!value) {
                        this.$message.error("名称不能为空")
                        return
                    }

                    // 中文字符最多16个 英文32个
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
                        return
                    }

                    this.editWorkspace(value)
                }).catch(() => {

                })
        },
        editWorkspace(newName) {
            if (this.space.name == newName) {
                return
            }

            // 本地先检查是否名称重复 
            this.$emit("onSpaceNameCheck", newName, this.index, async (ret) => {
                if (ret) {
                    this.$message.error("不能和已有的工作空间名称重复")
                    return
                }

                try {
                    // 发送请求修改名称
                    const {data:res} = await this.$axios.put("/api/workspace/name", {name: newName, id: this.space.id})
                    if (res.status) {
                        this.$message.error(res.message)
                        return
                    }

                    this.$message.success(res.message)
                    // 通知父组件修改名称
                    this.$emit("onSpaceNameModified",newName, this.index)
                } catch (error) {
                    this.$message.error('修改名称失败')
                }
            })
        }
    }
}
</script>

<style lang="scss" scoped>
@import "../assets/style/confirm.css";

.space-card {
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.03) 0%, rgba(255, 255, 255, 0.01) 100%);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 12px;
    transition: all 0.3s ease;
    backdrop-filter: blur(20px);
    position: relative;
    overflow: hidden;
    user-select: none;
    touch-action: pan-y;

    &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 3px;
        background: linear-gradient(90deg, transparent 0%, #64ffda 50%, transparent 100%);
        opacity: 0;
        transition: opacity 0.3s ease;
    }

    &.running {
        border-color: rgba(100, 255, 218, 0.3);
        
        &::before {
            opacity: 1;
        }
    }

    &.long-pressing {
        transform: scale(0.98);
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    }

    &.swipe-left {
        transform: translateX(-30px);
        
        .swipe-background {
            opacity: 1;
        }
    }

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        border-color: rgba(255, 255, 255, 0.15);
    }

    // 滑动删除背景
    .swipe-background {
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        width: 80px;
        background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transition: opacity 0.3s ease;
        z-index: 1;
        
        .swipe-action {
            display: flex;
            flex-direction: column;
            align-items: center;
            color: white;
            font-size: 12px;
            font-weight: 500;
            
            i {
                font-size: 20px;
                margin-bottom: 4px;
            }
        }
    }

    .card-content {
        display: flex;
        align-items: center;
        gap: 16px;
        position: relative;
        z-index: 2;
        background: inherit;
        border-radius: inherit;
        min-height: 60px;
    }

    // Logo区域
    .logo-section {
        flex-shrink: 0;
        
        .logo-container {
            position: relative;
            display: inline-block;
            
            .workspace-logo {
                width: 48px;
                height: 48px;
                border-radius: 8px;
                border: 2px solid rgba(255, 255, 255, 0.1);
                transition: all 0.3s ease;
            }
            
            .status-indicator {
                position: absolute;
                top: -3px;
                right: -3px;
                width: 14px;
                height: 14px;
                border-radius: 50%;
                background: #374151;
                border: 2px solid #1f2937;
                display: flex;
                align-items: center;
                justify-content: center;
                
                &.active {
                    background: #10b981;
                    
                    .status-dot {
                        width: 5px;
                        height: 5px;
                        background: #ffffff;
                        border-radius: 50%;
                        animation: pulse 2s infinite;
                    }
                }
            }
        }
    }

    // 信息区域
    .info-section {
        flex: 1;
        
        .workspace-info {
            .primary-info {
                margin-bottom: 6px;
                
                .workspace-name {
                    margin: 0 0 2px 0;
                    font-size: 16px;
                    font-weight: 600;
                    color: #ffffff;
                    line-height: 1.2;
                    display: -webkit-box;
                    -webkit-line-clamp: 1;
                    -webkit-box-orient: vertical;
                    overflow: hidden;
                }
                
                .workspace-status {
                    .status-text {
                        font-size: 11px;
                        font-weight: 500;
                        padding: 1px 6px;
                        border-radius: 10px;
                        background: rgba(107, 114, 128, 0.3);
                        color: #9ca3af;
                        
                        &.running {
                            background: rgba(16, 185, 129, 0.2);
                            color: #10b981;
                        }
                    }
                }
            }
            
            .secondary-info {
                display: flex;
                flex-direction: column;
                gap: 3px;
                
                .info-item {
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    font-size: 12px;
                    color: #9ca3af;
                    
                    i {
                        font-size: 12px;
                        color: #6b7280;
                        width: 14px;
                        flex-shrink: 0;
                    }
                    
                    .value {
                        flex: 1;
                        display: -webkit-box;
                        -webkit-line-clamp: 1;
                        -webkit-box-orient: vertical;
                        overflow: hidden;
                        line-height: 1.3;
                    }
                }
            }
        }
    }

    // 操作区域
    .actions-section {
        flex-shrink: 0;
        
        .action-buttons {
            display: flex;
            align-items: center;
            gap: 8px;
            
            .primary-actions {
                .primary-btn {
                    min-width: 70px;
                    height: 32px;
                    font-weight: 500;
                    border-radius: 6px;
                    font-size: 13px;
                    
                    .btn-text {
                        margin-left: 3px;
                    }
                    
                    &.enter-btn {
                        background: linear-gradient(135deg, #10b981 0%, #059669 100%);
                        border: none;
                        
                        &:hover {
                            background: linear-gradient(135deg, #059669 0%, #047857 100%);
                        }
                    }
                    
                    &.start-btn {
                        background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
                        border: none;
                        
                        &:hover {
                            background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
                        }
                    }
                }
            }
            
            .secondary-actions {
                display: flex;
                gap: 6px;
                
                .icon-btn {
                    width: 28px;
                    height: 28px;
                    border-radius: 50%;
                    background: rgba(255, 255, 255, 0.05);
                    border: 1px solid rgba(255, 255, 255, 0.1);
                    color: #9ca3af;
                    transition: all 0.3s ease;
                    
                    i {
                        font-size: 14px;
                    }
                    
                    &:hover {
                        background: rgba(255, 255, 255, 0.1);
                        color: #ffffff;
                        transform: scale(1.05);
                    }
                    
                    &.delete-btn {
                        color: #ef4444;
                        
                        &:hover {
                            background: rgba(239, 68, 68, 0.1);
                            color: #dc2626;
                        }
                    }
                }
            }
        }
    }

    // 长按菜单
    .context-menu {
        position: fixed;
        background: rgba(55, 65, 81, 0.95);
        backdrop-filter: blur(20px);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        padding: 8px 0;
        box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
        z-index: 1000;
        min-width: 180px;
        animation: contextMenuFadeIn 0.2s ease-out;
        
        .menu-item {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 10px 16px;
            color: #e5e7eb;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            
            i {
                font-size: 16px;
                width: 16px;
                color: #9ca3af;
            }
            
            &:hover {
                background: rgba(255, 255, 255, 0.1);
                color: #ffffff;
                
                i {
                    color: #64ffda;
                }
            }
            
            &.danger {
                color: #fca5a5;
                
                i {
                    color: #ef4444;
                }
                
                &:hover {
                    background: rgba(239, 68, 68, 0.1);
                    color: #ffffff;
                    
                    i {
                        color: #ef4444;
                    }
                }
            }
        }
    }

    .context-menu-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        z-index: 999;
        background: transparent;
    }

    // 移动端适配
    @media (max-width: 768px) {
        padding: 12px;
        margin-bottom: 8px;
        border-radius: 10px;

        .card-content {
            gap: 12px;
            align-items: center;
            min-height: 50px;
        }

        .logo-section {
            .logo-container .workspace-logo {
                width: 40px;
                height: 40px;
                border-radius: 8px;
            }
            
            .status-indicator {
                top: -2px;
                right: -2px;
                width: 12px;
                height: 12px;
                
                .status-dot {
                    width: 4px;
                    height: 4px;
                }
            }
        }

        .info-section {
            .workspace-info {
                .primary-info {
                    margin-bottom: 4px;
                    
                    .workspace-name {
                        font-size: 14px;
                    }
                }
                
                .secondary-info {
                    gap: 2px;
                    
                    .info-item {
                        font-size: 10px;
                        gap: 4px;
                        
                        i {
                            font-size: 10px;
                        }
                    }
                }
            }
        }

        .actions-section {
            .action-buttons {
                gap: 6px;
                
                .primary-actions {
                    .primary-btn {
                        min-width: 60px;
                        height: 28px;
                        font-size: 12px;
                        
                        .btn-text {
                            margin-left: 2px;
                        }
                    }
                }
                
                .secondary-actions {
                    gap: 6px;
                    
                    .icon-btn {
                        width: 32px;
                        height: 32px;
                        border-radius: 16px;
                        
                        i {
                            font-size: 14px;
                        }
                    }
                }
            }
        }
    }

    @media (max-width: 480px) {
        padding: 10px;
        margin-bottom: 6px;

        .card-content {
            gap: 10px;
            min-height: 45px;
        }

        .logo-section .logo-container .workspace-logo {
            width: 36px;
            height: 36px;
        }

        .info-section .workspace-info {
            .primary-info {
                margin-bottom: 3px;
                
                .workspace-name {
                    font-size: 13px;
                }
                
                .workspace-status .status-text {
                    font-size: 9px;
                    padding: 1px 4px;
                }
            }
            
            .secondary-info .info-item {
                font-size: 9px;
                gap: 3px;
                
                i {
                    font-size: 9px;
                }
            }
        }

        .actions-section .action-buttons {
            gap: 4px;
            
            .primary-actions .primary-btn {
                min-width: 55px;
                height: 26px;
                font-size: 11px;
            }
            
            .secondary-actions {
                gap: 4px;
                
                .icon-btn {
                    width: 30px;
                    height: 30px;
                    border-radius: 15px;
                    
                    i {
                        font-size: 13px;
                    }
                }
            }
        }
    }
}

@keyframes pulse {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.3;
    }
}

@keyframes contextMenuFadeIn {
    0% {
        opacity: 0;
        transform: scale(0.95) translateY(-10px);
    }
    100% {
        opacity: 1;
        transform: scale(1) translateY(0);
    }
}

.delete-confirm, .mobile-confirm-dialog {
    background-color: #363636;
    
    .el-message-box__header {
        border-bottom: 1px solid #4a4a4a;
    }
    
    .el-message-box__content {
        color: #e5e7eb;
    }
}

// 移动端对话框优化
@media (max-width: 768px) {
    .mobile-confirm-dialog {
        width: 90% !important;
        margin: 0 auto !important;
        
        .el-message-box__header {
            padding: 20px 20px 10px;
            
            .el-message-box__title {
                font-size: 16px;
            }
        }
        
        .el-message-box__content {
            padding: 10px 20px 20px;
            font-size: 14px;
        }
        
        .el-message-box__btns {
            padding: 0 20px 20px;
            
            .el-button {
                min-width: 80px;
                height: 36px;
            }
        }
    }
}
</style>
