import SwiftUI

struct ContentView: View {
    @EnvironmentObject var appState: AppState
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                // 标题
                HStack {
                    Image(systemName: "doc.text.magnifyingglass")
                        .font(.title)
                        .foregroundColor(.blue)
                    Text("YAML Monitor")
                        .font(.title)
                        .fontWeight(.bold)
                }
                .padding(.top)
                
                // 状态卡片
                StatusCardView()
                
                // 控制按钮
                ControlButtonsView()
                
                // 活动列表
                ActivityListView()
                
                Spacer()
            }
            .padding()
            .frame(minWidth: 600, minHeight: 500)
        }
        .onAppear {
            appState.checkPermissions()
        }
    }
}

struct StatusCardView: View {
    @EnvironmentObject var appState: AppState
    
    var body: some View {
        VStack(spacing: 12) {
            HStack {
                Text("系统状态")
                    .font(.headline)
                Spacer()
                Circle()
                    .fill(statusColor)
                    .frame(width: 12, height: 12)
            }
            
            VStack(alignment: .leading, spacing: 8) {
                StatusRow(title: "辅助功能权限", 
                         status: appState.hasAccessibilityPermission ? "已授权" : "未授权",
                         isGood: appState.hasAccessibilityPermission)
                
                StatusRow(title: "监控状态", 
                         status: appState.isMonitoring ? "运行中" : "已停止",
                         isGood: appState.isMonitoring)
                
                StatusRow(title: "连接状态", 
                         status: connectionStatusText,
                         isGood: appState.connectionStatus == .connected)
            }
        }
        .padding()
        .background(Color.gray.opacity(0.1))
        .cornerRadius(10)
    }
    
    private var statusColor: Color {
        if !appState.hasAccessibilityPermission {
            return .red
        } else if appState.isMonitoring {
            return .green
        } else {
            return .orange
        }
    }
    
    private var connectionStatusText: String {
        switch appState.connectionStatus {
        case .connected:
            return "已连接"
        case .disconnected:
            return "未连接"
        case .connecting:
            return "连接中..."
        }
    }
}

struct StatusRow: View {
    let title: String
    let status: String
    let isGood: Bool
    
    var body: some View {
        HStack {
            Text(title)
                .foregroundColor(.secondary)
            Spacer()
            Text(status)
                .foregroundColor(isGood ? .green : .red)
                .fontWeight(.medium)
        }
    }
}

struct ControlButtonsView: View {
    @EnvironmentObject var appState: AppState
    
    var body: some View {
        HStack(spacing: 16) {
            Button(action: {
                if appState.isMonitoring {
                    appState.stopMonitoring()
                } else {
                    appState.startMonitoring()
                }
            }) {
                HStack {
                    Image(systemName: appState.isMonitoring ? "stop.circle" : "play.circle")
                    Text(appState.isMonitoring ? "停止监控" : "开始监控")
                }
                .frame(minWidth: 120)
            }
            .buttonStyle(.borderedProminent)
            .disabled(!appState.hasAccessibilityPermission)
            
            Button(action: {
                appState.refreshActivities()
            }) {
                HStack {
                    Image(systemName: "arrow.clockwise")
                    Text("刷新")
                }
            }
            .buttonStyle(.bordered)
            
            Button(action: {
                PermissionManager.shared.checkAccessibilityPermission()
                DispatchQueue.main.asyncAfter(deadline: .now() + 1) {
                    appState.checkPermissions()
                }
            }) {
                HStack {
                    Image(systemName: "gear")
                    Text("权限设置")
                }
            }
            .buttonStyle(.bordered)
        }
    }
}

struct ActivityListView: View {
    @EnvironmentObject var appState: AppState
    
    var body: some View {
        VStack(alignment: .leading) {
            HStack {
                Text("最近活动")
                    .font(.headline)
                Spacer()
                Text("\(appState.activities.count) 条记录")
                    .foregroundColor(.secondary)
            }
            
            if appState.activities.isEmpty {
                VStack {
                    Image(systemName: "tray")
                        .font(.largeTitle)
                        .foregroundColor(.gray)
                    Text("暂无活动记录")
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity, minHeight: 100)
            } else {
                List(appState.activities.prefix(10)) { activity in
                    ActivityRowView(activity: activity)
                }
                .listStyle(.plain)
                .frame(maxHeight: 300)
            }
        }
    }
}

struct ActivityRowView: View {
    let activity: Activity
    
    var body: some View {
        HStack {
            Image(systemName: activity.type.icon)
                .foregroundColor(.blue)
                .frame(width: 20)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(activity.type.displayName)
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                if let content = activity.content, !content.isEmpty {
                    Text(content)
                        .lineLimit(1)
                } else if let appName = activity.appName {
                    Text(appName)
                        .lineLimit(1)
                }
            }
            
            Spacer()
            
            Text(formatDate(activity.timestamp))
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .padding(.vertical, 2)
    }
    
    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

#Preview {
    ContentView()
        .environmentObject(AppState())
}