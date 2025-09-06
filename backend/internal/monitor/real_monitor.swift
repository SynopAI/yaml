//
//  real_monitor.swift
//  YAML Real Monitor
//
//  Created for real keyboard and app monitoring on macOS
//

import Foundation
import Cocoa
import ApplicationServices

class RealMonitor {
    private var keyboardEventTap: CFMachPort?
    private var appObserver: NSWorkspace?
    private var isRunning = false
    private let outputPipe = Pipe()
    
    init() {
        self.appObserver = NSWorkspace.shared
    }
    
    func startMonitoring() {
        guard !isRunning else {
            print("[DEBUG] Monitor is already running")
            fflush(stdout)
            return
        }
        
        isRunning = true
        print("[DEBUG] Starting real monitoring...")
        fflush(stdout)
        print("[DEBUG] Process ID: \(ProcessInfo.processInfo.processIdentifier)")
        fflush(stdout)
        print("[DEBUG] Current user: \(NSUserName())")
        fflush(stdout)
        
        // 启动键盘监控
        print("[DEBUG] Initializing keyboard monitoring...")
        fflush(stdout)
        startKeyboardMonitoring()
        
        // 启动应用监控
        print("[DEBUG] Initializing app monitoring...")
        fflush(stdout)
        startAppMonitoring()
        
        print("[DEBUG] All monitors initialized, entering run loop...")
        fflush(stdout)
        // 保持程序运行
        RunLoop.main.run()
    }
    
    func stopMonitoring() {
        isRunning = false
        
        // 停止键盘监控
        if let eventTap = keyboardEventTap {
            CGEvent.tapEnable(tap: eventTap, enable: false)
            CFMachPortInvalidate(eventTap)
            keyboardEventTap = nil
        }
        
        print("Real monitoring stopped")
    }
    
    private func startKeyboardMonitoring() {
        print("[DEBUG] Checking accessibility permissions...")
        
        // 检查辅助功能权限
        let trusted = AXIsProcessTrustedWithOptions([
            kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String: true
        ] as CFDictionary)
        
        print("[DEBUG] Accessibility permission status: \(trusted)")
        
        guard trusted else {
            print("[ERROR] 需要辅助功能权限才能监控键盘输入")
            print("[ERROR] Please grant accessibility permission in System Preferences")
            return
        }
        
        print("[DEBUG] Creating keyboard event tap...")
        // 创建事件监听
        let eventMask = (1 << CGEventType.keyDown.rawValue)
        print("[DEBUG] Event mask: \(eventMask)")
        
        keyboardEventTap = CGEvent.tapCreate(
            tap: .cgSessionEventTap,
            place: .headInsertEventTap,
            options: .defaultTap,
            eventsOfInterest: CGEventMask(eventMask),
            callback: { (proxy, type, event, refcon) -> Unmanaged<CGEvent>? in
                let monitor = Unmanaged<RealMonitor>.fromOpaque(refcon!).takeUnretainedValue()
                monitor.handleKeyboardEvent(event: event)
                return Unmanaged.passUnretained(event)
            },
            userInfo: UnsafeMutableRawPointer(Unmanaged.passUnretained(self).toOpaque())
        )
        
        guard let eventTap = keyboardEventTap else {
            print("[ERROR] Failed to create keyboard event tap")
            print("[ERROR] This usually indicates insufficient permissions")
            return
        }
        
        print("[DEBUG] Event tap created successfully")
        print("[DEBUG] Adding to run loop...")
        
        let runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0)
        CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, .commonModes)
        CGEvent.tapEnable(tap: eventTap, enable: true)
        
        print("[SUCCESS] Keyboard monitoring started successfully")
    }
    
    private func startAppMonitoring() {
        print("[DEBUG] Setting up app monitoring...")
        
        guard let workspace = appObserver else {
            print("[ERROR] NSWorkspace not available")
            return
        }
        
        print("[DEBUG] NSWorkspace available, adding observers...")
        
        // 监听应用激活事件
        print("[DEBUG] Adding didActivateApplication observer...")
        workspace.notificationCenter.addObserver(
            forName: NSWorkspace.didActivateApplicationNotification,
            object: nil,
            queue: .main
        ) { [weak self] notification in
            print("[DEBUG] App activation event received")
            self?.handleAppActivation(notification: notification)
        }
        
        // 监听应用启动事件
        print("[DEBUG] Adding didLaunchApplication observer...")
        workspace.notificationCenter.addObserver(
            forName: NSWorkspace.didLaunchApplicationNotification,
            object: nil,
            queue: .main
        ) { [weak self] notification in
            print("[DEBUG] App launch event received")
            self?.handleAppLaunch(notification: notification)
        }
        
        // 监听应用终止事件
        print("[DEBUG] Adding didTerminateApplication observer...")
        workspace.notificationCenter.addObserver(
            forName: NSWorkspace.didTerminateApplicationNotification,
            object: nil,
            queue: .main
        ) { [weak self] notification in
            print("[DEBUG] App termination event received")
            self?.handleAppTermination(notification: notification)
        }
        
        print("[SUCCESS] App monitoring started successfully")
    }
    
    private func handleKeyboardEvent(event: CGEvent) {
        print("[DEBUG] Keyboard event received")
        
        guard isRunning else {
            print("[DEBUG] Monitor not running, ignoring keyboard event")
            return
        }
        
        // 获取按键信息
        let keyCode = event.getIntegerValueField(.keyboardEventKeycode)
        let flags = event.flags
        print("[DEBUG] Key code: \(keyCode), flags: \(flags.rawValue)")
        
        // 获取当前活跃应用
        let frontmostApp = NSWorkspace.shared.frontmostApplication
        let appName = frontmostApp?.localizedName ?? "Unknown"
        print("[DEBUG] Current app: \(appName)")
        
        // 尝试获取输入的字符
        var inputText = ""
        if let cgEventSource = CGEventSource(stateID: .hidSystemState) {
            let keyboardEvent = CGEvent(keyboardEventSource: cgEventSource, virtualKey: CGKeyCode(keyCode), keyDown: true)
            if let eventString = keyboardEvent?.keyboardGetUnicodeString(maxStringLength: 10, actualStringLength: nil, unicodeString: nil) {
                // 这里需要更复杂的字符提取逻辑
                inputText = "Key: \(keyCode)"
            }
        }
        print("[DEBUG] Input text: \(inputText)")
        
        // 输出键盘事件数据（JSON格式）
        let keyboardData: [String: Any] = [
            "type": "keyboard",
            "text": inputText,
            "app_name": appName,
            "timestamp": ISO8601DateFormatter().string(from: Date()),
            "key_code": keyCode,
            "modifiers": flags.rawValue
        ]
        
        print("[DEBUG] Outputting keyboard event data")
        outputEvent(data: keyboardData)
    }
    
    private func handleAppActivation(notification: Notification) {
        print("[DEBUG] App activation notification received")
        
        guard isRunning else {
            print("[DEBUG] Monitor not running, ignoring app activation")
            return
        }
        
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            let appName = app.localizedName ?? "Unknown"
            let bundleId = app.bundleIdentifier ?? "Unknown"
            print("[DEBUG] App activated: \(appName) (\(bundleId))")
            
            let appData: [String: Any] = [
                "type": "app_activation",
                "app_name": appName,
                "bundle_id": bundleId,
                "timestamp": ISO8601DateFormatter().string(from: Date()),
                "pid": app.processIdentifier
            ]
            
            print("[DEBUG] Outputting app activation data")
            outputEvent(data: appData)
        } else {
            print("[DEBUG] No app info in activation notification")
        }
    }
    
    private func handleAppLaunch(notification: Notification) {
        guard isRunning else { return }
        
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            let appData: [String: Any] = [
                "type": "app_launch",
                "app_name": app.localizedName ?? "Unknown",
                "bundle_id": app.bundleIdentifier ?? "Unknown",
                "timestamp": ISO8601DateFormatter().string(from: Date()),
                "pid": app.processIdentifier
            ]
            
            outputEvent(data: appData)
        }
    }
    
    private func handleAppTermination(notification: Notification) {
        guard isRunning else { return }
        
        if let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication {
            let appData: [String: Any] = [
                "type": "app_termination",
                "app_name": app.localizedName ?? "Unknown",
                "bundle_id": app.bundleIdentifier ?? "Unknown",
                "timestamp": ISO8601DateFormatter().string(from: Date()),
                "pid": app.processIdentifier
            ]
            
            outputEvent(data: appData)
        }
    }
    
    private func outputEvent(data: [String: Any]) {
        print("[DEBUG] Preparing to output event data: \(data)")
        fflush(stdout)
        
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: data, options: [])
            if let jsonString = String(data: jsonData, encoding: .utf8) {
                print("[DEBUG] JSON serialized successfully: \(jsonString)")
                fflush(stdout)
                print("YAML_EVENT: \(jsonString)")
                fflush(stdout)
                print("[DEBUG] Event output completed")
                fflush(stdout)
            } else {
                print("[ERROR] Failed to convert JSON data to string")
                fflush(stdout)
            }
        } catch {
            print("[ERROR] Error serializing event data: \(error)")
            fflush(stdout)
        }
    }
}

// 主程序入口
class MonitorApp {
    static func main() {
        let monitor = RealMonitor()
        
        // 处理信号
        signal(SIGINT) { _ in
            print("\nReceived SIGINT, stopping monitor...")
            exit(0)
        }
        
        signal(SIGTERM) { _ in
            print("\nReceived SIGTERM, stopping monitor...")
            exit(0)
        }
        
        print("YAML Real Monitor starting...")
        monitor.startMonitoring()
    }
}

// 启动监控
MonitorApp.main()