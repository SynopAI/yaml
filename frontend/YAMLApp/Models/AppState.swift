import Foundation
import Combine

class AppState: ObservableObject {
    @Published var isMonitoring = false
    @Published var hasAccessibilityPermission = false
    @Published var activities: [Activity] = []
    @Published var connectionStatus: ConnectionStatus = .disconnected
    
    private var cancellables = Set<AnyCancellable>()
    private let apiService = APIService()
    
    enum ConnectionStatus {
        case connected
        case disconnected
        case connecting
    }
    
    init() {
        checkPermissions()
        setupPeriodicUpdates()
    }
    
    func checkPermissions() {
        hasAccessibilityPermission = PermissionManager.shared.hasAccessibilityPermission()
    }
    
    func startMonitoring() {
        guard hasAccessibilityPermission else {
            print("No accessibility permission")
            return
        }
        
        connectionStatus = .connecting
        
        apiService.startMonitoring { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success:
                    self?.isMonitoring = true
                    self?.connectionStatus = .connected
                case .failure(let error):
                    print("Failed to start monitoring: \(error)")
                    self?.connectionStatus = .disconnected
                }
            }
        }
    }
    
    func stopMonitoring() {
        apiService.stopMonitoring { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success:
                    self?.isMonitoring = false
                    self?.connectionStatus = .disconnected
                case .failure(let error):
                    print("Failed to stop monitoring: \(error)")
                }
            }
        }
    }
    
    func refreshActivities() {
        apiService.getActivities { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success(let activities):
                    self?.activities = activities
                case .failure(let error):
                    print("Failed to fetch activities: \(error)")
                }
            }
        }
    }
    
    private func setupPeriodicUpdates() {
        Timer.publish(every: 5.0, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                if self?.isMonitoring == true {
                    self?.refreshActivities()
                }
            }
            .store(in: &cancellables)
    }
}