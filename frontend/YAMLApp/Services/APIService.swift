import Foundation

class APIService {
    private let baseURL = "http://localhost:8080/api/v1"
    private let session = URLSession.shared
    
    func startMonitoring(completion: @escaping (Result<Void, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/monitor/start") else {
            completion(.failure(APIError.invalidURL))
            return
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        session.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                completion(.failure(APIError.serverError))
                return
            }
            
            completion(.success(()))
        }.resume()
    }
    
    func stopMonitoring(completion: @escaping (Result<Void, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/monitor/stop") else {
            completion(.failure(APIError.invalidURL))
            return
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        session.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                completion(.failure(APIError.serverError))
                return
            }
            
            completion(.success(()))
        }.resume()
    }
    
    func getActivities(limit: Int = 50, completion: @escaping (Result<[Activity], Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/activities?limit=\(limit)") else {
            completion(.failure(APIError.invalidURL))
            return
        }
        
        session.dataTask(with: url) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            
            guard let data = data else {
                completion(.failure(APIError.noData))
                return
            }
            
            do {
                let decoder = JSONDecoder()
                decoder.dateDecodingStrategy = .iso8601
                let response = try decoder.decode(ActivitiesResponse.self, from: data)
                completion(.success(response.activities))
            } catch {
                completion(.failure(error))
            }
        }.resume()
    }
    
    func getMonitorStatus(completion: @escaping (Result<MonitorStatus, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/monitor/status") else {
            completion(.failure(APIError.invalidURL))
            return
        }
        
        session.dataTask(with: url) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            
            guard let data = data else {
                completion(.failure(APIError.noData))
                return
            }
            
            do {
                let status = try JSONDecoder().decode(MonitorStatus.self, from: data)
                completion(.success(status))
            } catch {
                completion(.failure(error))
            }
        }.resume()
    }
}

enum APIError: Error, LocalizedError {
    case invalidURL
    case noData
    case serverError
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "无效的URL"
        case .noData:
            return "没有数据"
        case .serverError:
            return "服务器错误"
        }
    }
}