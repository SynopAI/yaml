import Foundation

struct Activity: Codable, Identifiable {
    let id: Int64
    let type: ActivityType
    let content: String?
    let appName: String?
    let windowTitle: String?
    let url: String?
    let timestamp: Date
    let duration: Int64
    
    enum ActivityType: String, Codable, CaseIterable {
        case keyboard = "keyboard"
        case app = "app"
        case web = "web"
        case click = "click"
        
        var displayName: String {
            switch self {
            case .keyboard:
                return "键盘输入"
            case .app:
                return "应用切换"
            case .web:
                return "网页浏览"
            case .click:
                return "点击操作"
            }
        }
        
        var icon: String {
            switch self {
            case .keyboard:
                return "keyboard"
            case .app:
                return "app.badge"
            case .web:
                return "globe"
            case .click:
                return "cursorarrow.click"
            }
        }
    }
}

struct ActivitiesResponse: Codable {
    let activities: [Activity]
    let count: Int
}

struct KeyboardInput: Codable {
    let text: String
    let appName: String?
    let timestamp: Date
}

struct APIResponse: Codable {
    let message: String
}

struct MonitorStatus: Codable {
    let status: [String: Bool]
    let running: Bool
}