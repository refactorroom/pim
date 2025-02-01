// Basic JSON logging
data := map[string]interface{}{
    "name": "John",
    "age": 30,
    "isActive": true,
}
console.Json(data)
console.JsonWithLabel("User Data", data)

// Request logging
console.JsonRequest("POST", "/api/users", data, map[string]interface{}{
    "headers": map[string]string{
        "Content-Type": "application/json",
    },
    "labels": []string{"api", "create"},
})

// Response logging
response := map[string]interface{}{
    "id": "123",
    "status": "created",
}
console.JsonResponse(201, response, map[string]interface{}{
    "duration": "125ms",
    "labels": []string{"success"},
})