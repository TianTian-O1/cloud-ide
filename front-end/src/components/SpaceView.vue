
<template>
    <div class="space-view-wrapper">
        <!-- 页面标题 -->
        <div class="page-header">
            <div class="header-content">
                <h1 class="page-title">我的工作空间</h1>
                <div class="workspace-count" v-if="dataLoaded">
                    共 {{ spaces.length }} 个空间
                </div>
            </div>
        </div>

        <!-- 加载状态 -->
        <div v-if="!dataLoaded" class="loading-container">
            <div class="loading-skeleton" v-for="n in 3" :key="n">
                <div class="skeleton-logo"></div>
                <div class="skeleton-content">
                    <div class="skeleton-line short"></div>
                    <div class="skeleton-line medium"></div>
                    <div class="skeleton-line long"></div>
                </div>
                <div class="skeleton-actions">
                    <div class="skeleton-btn" v-for="i in 4" :key="i"></div>
                </div>
            </div>
        </div>

        <!-- 工作空间列表 -->
        <div v-else-if="spaces.length > 0" class="spaces-container">
            <SpaceCard 
                v-for="(item, i) in spaces"  
                :key="item.id" 
                :space="item" 
                :index="i" 
                @onDeleteSpace="deleteElement"
                @onStopSpace="setSpaceStatus" 
                @onStartSpace="setSpaceStatus"
                @onSpaceNameCheck="checkSpaceName" 
                @onSpaceNameModified="modifySpaceName"
                @click="startSpace(item.id)"
            />
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-state">
            <div class="empty-icon">
                <i class="el-icon-folder-opened"></i>
            </div>
            <h3 class="empty-title">还没有工作空间</h3>
            <p class="empty-description">
                创建您的第一个工作空间，开始编程之旅
            </p>
            <el-button 
                type="primary" 
                class="create-workspace-btn"
                @click="goToTemplates"
            >
                <i class="el-icon-plus"></i>
                创建工作空间
            </el-button>
        </div>
    </div>
</template>

<script>
import SpaceCard from './SpaceCard.vue'

export default {
    components: {
        SpaceCard,
    },
    data() {
        return {
            dataLoaded: false,
            spaces: [],
        }
    },
    methods: {
        async getAllSpaces() {
            try {
            const {data:res} = await this.$axios.get("/api/workspace/list")
            if (res.status) {
                this.$message.error(res.message)
                return
            }
                this.spaces = res.data || []
            } catch (error) {
                console.error('获取工作空间列表失败:', error)
                this.$message.error('获取工作空间列表失败')
            }
        },

        async startSpace(id) {
            // 启动工作空间的逻辑
        },
        
        deleteElement(index) {
            this.spaces.splice(index, 1)
        },
        
        setSpaceStatus(index, status) {
            this.spaces[index].running_status = status
        },
        
        checkSpaceName(name, index, callback) {
            // 检查是否有工作空间的名称和name相同
            for (let i = 0; i < this.spaces.length; i++) {
                const ele = this.spaces[i]
                if (i == index) {
                    continue
                }
                
                if (ele.name === name) {
                    callback(true)
                    return
                }
            }

            callback(false)
        },
        
        modifySpaceName(newName, index) {
            this.spaces[index].name = newName
        },

        goToTemplates() {
            this.$router.push('/dash/templates')
        }
    },
    
    async mounted() {
        await this.getAllSpaces()
        this.dataLoaded = true
    }
}
</script>

<style lang="scss" scoped>
.space-view-wrapper {
    min-height: 100vh;
    background: linear-gradient(135deg, #1e1f26 0%, #2a2d3a 100%);
    padding: 20px 25px;

    @media (max-width: 768px) {
        padding: 15px 12px;
    }

    @media (max-width: 480px) {
        padding: 12px 8px;
    }
}

// 页面头部
.page-header {
    margin-bottom: 30px;

    .header-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        flex-wrap: wrap;
        gap: 12px;
    }

    .page-title {
        margin: 0;
        font-size: 28px;
        font-weight: 600;
        background: linear-gradient(135deg, #64ffda 0%, #4db6ac 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;

        @media (max-width: 768px) {
            font-size: 24px;
        }

        @media (max-width: 480px) {
            font-size: 20px;
        }
    }

    .workspace-count {
        color: #a1a1aa;
        font-size: 14px;
        font-weight: 500;
        background: rgba(255, 255, 255, 0.05);
        padding: 6px 12px;
        border-radius: 20px;
        backdrop-filter: blur(10px);

        @media (max-width: 768px) {
            font-size: 12px;
            padding: 4px 8px;
        }
    }

    @media (max-width: 768px) {
        margin-bottom: 20px;

        .header-content {
            flex-direction: column;
            align-items: flex-start;
        }
    }
}

// 加载状态
.loading-container {
    .loading-skeleton {
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid rgba(255, 255, 255, 0.06);
        border-radius: 12px;
        padding: 20px;
        margin-bottom: 16px;
        display: flex;
        align-items: center;
        gap: 16px;

        .skeleton-logo {
            width: 48px;
            height: 48px;
            background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
            background-size: 200% 100%;
            animation: shimmer 1.5s infinite;
            border-radius: 8px;
            flex-shrink: 0;
        }

        .skeleton-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            gap: 8px;

            .skeleton-line {
                height: 14px;
                background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
                background-size: 200% 100%;
                animation: shimmer 1.5s infinite;
                border-radius: 4px;

                &.short { width: 40%; }
                &.medium { width: 60%; }
                &.long { width: 80%; }
            }
        }

        .skeleton-actions {
            display: flex;
            gap: 8px;
            flex-shrink: 0;

            .skeleton-btn {
                width: 40px;
                height: 40px;
                background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
                background-size: 200% 100%;
                animation: shimmer 1.5s infinite;
                border-radius: 50%;
            }
        }

        @media (max-width: 768px) {
            flex-direction: column;
            align-items: stretch;
            gap: 12px;
            padding: 16px;

            .skeleton-logo {
                align-self: center;
                width: 40px;
                height: 40px;
            }

            .skeleton-actions {
                justify-content: center;
            }
        }
    }
}

@keyframes shimmer {
    0% { background-position: -200% 0; }
    100% { background-position: 200% 0; }
}

// 工作空间容器
.spaces-container {
    display: flex;
    flex-direction: column;
    gap: 16px;

    @media (max-width: 768px) {
        gap: 12px;
    }

    @media (max-width: 480px) {
        gap: 10px;
    }
}

// 空状态
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 60px 20px;
    margin-top: 40px;

    .empty-icon {
        margin-bottom: 24px;
        
        i {
            font-size: 80px;
            color: #4b5563;
        }
    }

    .empty-title {
        margin: 0 0 12px 0;
        font-size: 24px;
        font-weight: 600;
        color: #e5e7eb;
    }

    .empty-description {
        margin: 0 0 32px 0;
        font-size: 16px;
        color: #9ca3af;
        max-width: 400px;
        line-height: 1.5;
    }

    .create-workspace-btn {
        padding: 12px 24px;
        font-size: 16px;
        font-weight: 500;
        border-radius: 8px;
        background: linear-gradient(135deg, #64ffda 0%, #4db6ac 100%);
        border: none;
        color: #1a202c;
        
        &:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(100, 255, 218, 0.3);
        }

        i {
            margin-right: 8px;
        }
    }

    @media (max-width: 768px) {
        padding: 40px 16px;
        margin-top: 20px;

        .empty-icon i {
            font-size: 60px;
        }

        .empty-title {
            font-size: 20px;
        }

        .empty-description {
            font-size: 14px;
        }

        .create-workspace-btn {
            padding: 10px 20px;
            font-size: 14px;
        }
    }

    @media (max-width: 480px) {
        padding: 30px 12px;

        .empty-icon i {
            font-size: 50px;
        }

        .empty-title {
    font-size: 18px;
        }

        .empty-description {
            font-size: 13px;
}
    }
}    
</style>