package main

type Config struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

type Task struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	LongDescription string      `json:"longDescription"`
	Project         string      `json:"project"`
	ProjectUUID     interface{} `json:"projectUuid"`
	Timestamp       interface{} `json:"timestamp"`
	Status          int         `json:"status"`
	Archived        bool        `json:"archived"`
	Periodicity     interface{} `json:"periodicity"`
	Username        string      `json:"username"`
	Workspace       string      `json:"workspace"`
	WorkspaceUUID   interface{} `json:"workspaceUuid"`
	Position        float64     `json:"position"`
	ParentID        interface{} `json:"parentId"`
}

type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Username  string `json:"username"`
	Workspace string `json:"workspace"`
}

type Data struct {
	Tasks    []Task    `json:"simplanner-tasks"`
	Projects []Project `json:"simplanner-projects"`
}

type AuthResponse struct {
	Token               string `json:"token"`
	IsTemporaryPassword bool   `json:"is_temporary_password"`
}

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  int    `json:"status"`
}
