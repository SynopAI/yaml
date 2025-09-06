// 前端配置文件
const CONFIG = {
    // API 配置
    API: {
        BASE_URL: 'http://localhost:8080/api/v1',
        TIMEOUT: 30000, // 30秒超时
    },
    
    // 自动刷新配置
    AUTO_REFRESH: {
        ENABLED: true,
        INTERVAL: 5000, // 5秒间隔
    },
    
    // 数据显示配置
    DISPLAY: {
        DEFAULT_LIMIT: 20,
        MAX_LIMIT: 100,
        ACTIVITY_LIMIT: 20,
        KEYBOARD_LIMIT: 20,
        SUMMARY_LIMIT: 10,
    },
    
    // UI 配置
    UI: {
        THEME: 'light', // light, dark
        LANGUAGE: 'zh-CN', // zh-CN, en-US
    }
};

// 导出配置
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CONFIG;
} else if (typeof window !== 'undefined') {
    window.CONFIG = CONFIG;
}