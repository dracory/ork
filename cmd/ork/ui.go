package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dracory/ork/vault"
)

// apiResponse represents a JSON API response
type apiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (r apiResponse) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return `{"status":"error","message":"internal server error"}`
	}
	return string(b)
}

// apiSuccess creates a success response
func apiSuccess(message string) apiResponse {
	return apiResponse{
		Status:  "success",
		Message: message,
	}
}

// apiSuccessWithData creates a success response with data
func apiSuccessWithData(message string, data interface{}) apiResponse {
	return apiResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// apiError creates an error response
func apiError(message string) apiResponse {
	return apiResponse{
		Status:  "error",
		Message: message,
	}
}

// startUIServer starts the HTTP server for the vault UI
func startUIServer(vaultPath string, address string) error {
	ui := &vaultUI{
		vaultPath: vaultPath,
	}

	http.HandleFunc("/", ui.handleRequest)

	fmt.Printf("Server listening on http://%s\n", address)
	return http.ListenAndServe(address, nil)
}

type vaultUI struct {
	vaultPath string
}

func (u *vaultUI) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	action := r.URL.Query().Get("a")

	switch action {
	case "login", "keys", "key-add", "key-update", "key-remove":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(apiError("Method not allowed").String()))
			return
		}
		switch action {
		case "login":
			u.handleLogin(w, r)
		case "keys":
			u.handleKeys(w, r)
		case "key-add":
			u.handleKeyAdd(w, r)
		case "key-update":
			u.handleKeyUpdate(w, r)
		case "key-remove":
			u.handleKeyRemove(w, r)
		}
	default:
		// Serve HTML page
		u.handlePage(w, r)
	}
}

func (u *vaultUI) handlePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(u.getHTML()))
}

func (u *vaultUI) handleLogin(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if password == "" {
		w.Write([]byte(apiError("Password is required").String()))
		return
	}

	// Open vault
	v, err := vault.Open(u.vaultPath, password)
	if err != nil {
		w.Write([]byte(apiError("Invalid password").String()))
		return
	}
	defer v.Close()

	keys := v.KeyList()
	keysMap := make(map[string]string)
	for _, key := range keys {
		value, _ := v.KeyGet(key)
		keysMap[key] = value
	}

	w.Write([]byte(apiSuccessWithData("Login successful", map[string]interface{}{
		"keys": keysMap,
	}).String()))
}

func (u *vaultUI) handleKeys(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if password == "" {
		w.Write([]byte(apiError("Password is required").String()))
		return
	}

	v, err := vault.Open(u.vaultPath, password)
	if err != nil {
		w.Write([]byte(apiError("Failed to open vault").String()))
		return
	}
	defer v.Close()

	keys := v.KeyList()
	keysMap := make(map[string]string)
	for _, key := range keys {
		value, _ := v.KeyGet(key)
		keysMap[key] = value
	}

	w.Write([]byte(apiSuccessWithData("Keys retrieved", map[string]interface{}{
		"keys": keysMap,
	}).String()))
}

func (u *vaultUI) handleKeyAdd(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	key := r.FormValue("key")
	value := r.FormValue("value")

	if password == "" {
		w.Write([]byte(apiError("Password is required").String()))
		return
	}
	if key == "" {
		w.Write([]byte(apiError("Key is required").String()))
		return
	}

	v, err := vault.Open(u.vaultPath, password)
	if err != nil {
		w.Write([]byte(apiError("Failed to open vault").String()))
		return
	}

	if v.KeyExists(key) {
		v.Close()
		w.Write([]byte(apiError("Key already exists").String()))
		return
	}

	if err := v.KeySet(key, value); err != nil {
		v.Close()
		w.Write([]byte(apiError("Failed to set key").String()))
		return
	}

	if err := v.Close(); err != nil {
		w.Write([]byte(apiError("Failed to save vault").String()))
		return
	}

	w.Write([]byte(apiSuccess("Key added successfully").String()))
}

func (u *vaultUI) handleKeyUpdate(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	key := r.FormValue("key")
	value := r.FormValue("value")

	if password == "" {
		w.Write([]byte(apiError("Password is required").String()))
		return
	}
	if key == "" {
		w.Write([]byte(apiError("Key is required").String()))
		return
	}

	v, err := vault.Open(u.vaultPath, password)
	if err != nil {
		w.Write([]byte(apiError("Failed to open vault").String()))
		return
	}

	if !v.KeyExists(key) {
		v.Close()
		w.Write([]byte(apiError("Key does not exist").String()))
		return
	}

	if err := v.KeySet(key, value); err != nil {
		v.Close()
		w.Write([]byte(apiError("Failed to update key").String()))
		return
	}

	if err := v.Close(); err != nil {
		w.Write([]byte(apiError("Failed to save vault").String()))
		return
	}

	w.Write([]byte(apiSuccess("Key updated successfully").String()))
}

func (u *vaultUI) handleKeyRemove(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	key := r.FormValue("key")

	if password == "" {
		w.Write([]byte(apiError("Password is required").String()))
		return
	}
	if key == "" {
		w.Write([]byte(apiError("Key is required").String()))
		return
	}

	v, err := vault.Open(u.vaultPath, password)
	if err != nil {
		w.Write([]byte(apiError("Failed to open vault").String()))
		return
	}

	if !v.KeyExists(key) {
		v.Close()
		w.Write([]byte(apiError("Key does not exist").String()))
		return
	}

	if err := v.KeyDelete(key); err != nil {
		v.Close()
		w.Write([]byte(apiError("Failed to delete key").String()))
		return
	}

	if err := v.Close(); err != nil {
		w.Write([]byte(apiError("Failed to save vault").String()))
		return
	}

	w.Write([]byte(apiSuccess("Key removed successfully").String()))
}

func (u *vaultUI) getCSS() string {
	return `<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/notiflix@3.2.6/dist/notiflix-3.2.6.min.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 50px auto; padding: 20px; }
        .card { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); padding: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 8px; font-weight: 500; }
        input, textarea { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        textarea { min-height: 100px; resize: vertical; }
        button { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; font-weight: 500; }
        .btn-primary { background: #007bff; color: white; }
        .btn-primary:hover { background: #0056b3; }
        .btn-success { background: #28a745; color: white; }
        .btn-success:hover { background: #1e7e34; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-danger:hover { background: #c82333; }
        .btn-secondary { background: #6c757d; color: white; }
        .btn-secondary:hover { background: #545b62; }
        .alert { padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        .alert-danger { background: #f8d7da; color: #721c24; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8f9fa; font-weight: 600; }
        pre { background: #f8f9fa; padding: 8px; border-radius: 4px; white-space: pre-wrap; word-wrap: break-word; }
        .modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000; }
        .modal.show { display: flex; align-items: center; justify-content: center; }
        .modal-content { background: white; border-radius: 8px; padding: 30px; max-width: 500px; width: 90%; }
        .modal-header { font-size: 20px; font-weight: 600; margin-bottom: 20px; }
        .modal-footer { display: flex; justify-content: space-between; margin-top: 20px; }
        .hidden { display: none !important; }
        .login-container { display: flex; align-items: center; justify-content: center; min-height: 100vh; }
        .login-card { width: 100%; max-width: 400px; }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
        .header h1 { margin: 0; }
        .actions { display: flex; gap: 10px; }
        .actions button { padding: 6px 12px; font-size: 12px; }
    </style>`
}

func (u *vaultUI) getJS(vaultPathJSON string) string {
	return `<script>
        const vaultPath = ` + vaultPathJSON + `;
        window.addEventListener('DOMContentLoaded', function() {
            try {
                Notiflix.Notify.init({
                    position: 'right-top',
                    distance: '10px',
                    timeout: 3000,
                    showOnlyTheLastOne: true
                });
                const app = Vue.createApp({
                    data() {
                        return {
                            pageKeysShow: false,
                            pageLoginShow: true,
                            keys: {},
                            vaultPassword: '',
                            keyAddModalVisible: false,
                            keyUpdateModalVisible: false,
                            keyRemoveModalVisible: false,
                            keyAddForm: { key: '', value: '', errorMessage: '' },
                            keyUpdateForm: { key: '', value: '', errorMessage: '' },
                            keyRemoveForm: { key: '', errorMessage: '' },
                            loginForm: { password: '', errorMessage: '' }
                        }
                    },
                    methods: {
                        login() {
                            this.loginForm.errorMessage = '';
                            const formData = new FormData();
                            formData.append('password', this.loginForm.password);
                            fetch('?a=login', { method: 'POST', body: formData })
                                .then(r => r.json())
                                .then(response => {
                                    if (response.status !== 'success') {
                                        this.loginForm.errorMessage = response.message;
                                        Notiflix.Notify.failure('Login failed: ' + response.message);
                                        return;
                                    }
                                    this.pageKeysShow = true;
                                    this.pageLoginShow = false;
                                    this.keys = response.data.keys;
                                    this.vaultPassword = this.loginForm.password;
                                    Notiflix.Notify.success('Login successful');
                                })
                                .catch(() => {
                                    this.loginForm.errorMessage = 'Login failed';
                                    Notiflix.Notify.failure('Login failed');
                                });
                        },
                        keysList() {
                            const formData = new FormData();
                            formData.append('password', this.vaultPassword);
                            fetch('?a=keys', { method: 'POST', body: formData })
                                .then(r => r.json())
                                .then(response => {
                                    if (response.status !== 'success') {
                                        this.pageKeysShow = false;
                                        this.pageLoginShow = true;
                                        return;
                                    }
                                    this.keys = response.data.keys;
                                })
                                .catch(() => {});
                        },
                        keyAddModal() {
                            this.keyAddForm = { key: '', value: '', errorMessage: '' };
                            this.keyAddModalVisible = true;
                        },
                        keyAddModalClose() {
                            this.keyAddModalVisible = false;
                        },
                        keyAdd() {
                            const formData = new FormData();
                            formData.append('password', this.vaultPassword);
                            formData.append('key', this.keyAddForm.key);
                            formData.append('value', this.keyAddForm.value);
                            fetch('?a=key-add', { method: 'POST', body: formData })
                                .then(r => r.json())
                                .then(response => {
                                    if (response.status !== 'success') {
                                        this.keyAddForm.errorMessage = response.message;
                                        Notiflix.Notify.failure('Failed to add key: ' + response.message);
                                        return;
                                    }
                                    this.keyAddModalClose();
                                    this.keysList();
                                    Notiflix.Notify.success('Key "' + this.keyAddForm.key + '" added successfully');
                                })
                                .catch(() => {
                                    this.keyAddForm.errorMessage = 'Adding key failed';
                                    Notiflix.Notify.failure('Failed to add key');
                                });
                        },
                        keyUpdateModalShow(key) {
                            this.keyUpdateForm = { key: key, value: this.keys[key], errorMessage: '' };
                            this.keyUpdateModalVisible = true;
                        },
                        keyUpdateModalClose() {
                            this.keyUpdateModalVisible = false;
                        },
                        keyUpdate() {
                            const formData = new FormData();
                            formData.append('password', this.vaultPassword);
                            formData.append('key', this.keyUpdateForm.key);
                            formData.append('value', this.keyUpdateForm.value);
                            fetch('?a=key-update', { method: 'POST', body: formData })
                                .then(r => r.json())
                                .then(response => {
                                    if (response.status !== 'success') {
                                        this.keyUpdateForm.errorMessage = response.message;
                                        Notiflix.Notify.failure('Failed to update key: ' + response.message);
                                        return;
                                    }
                                    this.keyUpdateModalClose();
                                    this.keysList();
                                    Notiflix.Notify.success('Key "' + this.keyUpdateForm.key + '" updated successfully');
                                })
                                .catch(() => {
                                    this.keyUpdateForm.errorMessage = 'Updating key failed';
                                    Notiflix.Notify.failure('Failed to update key');
                                });
                        },
                        keyRemoveModalShow(key) {
                            this.keyRemoveForm = { key: key, errorMessage: '' };
                            this.keyRemoveModalVisible = true;
                        },
                        keyRemoveModalClose() {
                            this.keyRemoveModalVisible = false;
                        },
                        keyRemove() {
                            const formData = new FormData();
                            formData.append('password', this.vaultPassword);
                            formData.append('key', this.keyRemoveForm.key);
                            fetch('?a=key-remove', { method: 'POST', body: formData })
                                .then(r => r.json())
                                .then(response => {
                                    if (response.status !== 'success') {
                                        this.keyRemoveForm.errorMessage = response.message;
                                        Notiflix.Notify.failure('Failed to remove key: ' + response.message);
                                        return;
                                    }
                                    this.keyRemoveModalClose();
                                    this.keysList();
                                    Notiflix.Notify.success('Key "' + this.keyRemoveForm.key + '" removed successfully');
                                })
                                .catch(() => {
                                    this.keyRemoveForm.errorMessage = 'Removing key failed';
                                    Notiflix.Notify.failure('Failed to remove key');
                                });
                        }
                    }
                });
                app.mount('#app');
                console.log('Vue app mounted successfully');
            } catch (error) {
                console.error('Failed to mount Vue app:', error);
            }
        });
    </script>`
}

func (u *vaultUI) getHTML() string {
	vaultPathJSON, err := json.Marshal(u.vaultPath)
	if err != nil {
		vaultPathJSON = []byte(`""`)
	}
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ork Vault UI</title>
    ` + u.getCSS() + `
    <script src="https://cdn.jsdelivr.net/npm/vue@3/dist/vue.global.prod.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/notiflix@3.2.6/dist/notiflix-3.2.6.min.js"></script>
</head>
<body>
    <div id="app">
        <!-- Login Page -->
        <div class="login-container" v-if="pageLoginShow">
            <div class="card login-card">
                <h1 style="text-align: center; margin-bottom: 30px;">🔐 Ork Vault</h1>
                <div class="alert alert-danger" v-if="loginForm.errorMessage">{{ loginForm.errorMessage }}</div>
                <div class="form-group">
                    <label>Password</label>
                    <input type="password" v-model="loginForm.password" @keyup.enter="login">
                </div>
                <button class="btn-primary" style="width: 100%" @click="login">Login</button>
            </div>
        </div>

        <!-- Keys Page -->
        <div class="container" v-if="pageKeysShow">
            <div class="card">
                <div class="header">
                    <h1>Keys</h1>
                    <button class="btn-success" @click="keyAddModal">+ New Key</button>
                </div>
                <table>
                    <thead>
                        <tr>
                            <th>Key</th>
                            <th>Value</th>
                            <th style="width: 150px;">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="key in Object.keys(keys)" :key="key">
                            <td>{{ key }}</td>
                            <td><pre>{{ keys[key].substring(0, 100) }}{{ keys[key].length > 100 ? '...' : '' }}</pre></td>
                            <td class="actions">
                                <button class="btn-primary" @click="keyUpdateModalShow(key)">Edit</button>
                                <button class="btn-danger" @click="keyRemoveModalShow(key)">Delete</button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Add Key Modal -->
        <div class="modal" :class="{ show: keyAddModalVisible }">
            <div class="modal-content">
                <div class="modal-header">Add Key</div>
                <div class="alert alert-danger" v-if="keyAddForm.errorMessage">{{ keyAddForm.errorMessage }}</div>
                <div class="form-group">
                    <label>Key</label>
                    <input type="text" v-model="keyAddForm.key">
                </div>
                <div class="form-group">
                    <label>Value</label>
                    <textarea v-model="keyAddForm.value"></textarea>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" @click="keyAddModalClose">Cancel</button>
                    <button class="btn-primary" @click="keyAdd">Save</button>
                </div>
            </div>
        </div>

        <!-- Update Key Modal -->
        <div class="modal" :class="{ show: keyUpdateModalVisible }">
            <div class="modal-content">
                <div class="modal-header">Update Key</div>
                <div class="alert alert-danger" v-if="keyUpdateForm.errorMessage">{{ keyUpdateForm.errorMessage }}</div>
                <div class="form-group">
                    <label>Key</label>
                    <input type="text" v-model="keyUpdateForm.key" readonly style="background: #f5f5f5;">
                </div>
                <div class="form-group">
                    <label>Value</label>
                    <textarea v-model="keyUpdateForm.value"></textarea>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" @click="keyUpdateModalClose">Cancel</button>
                    <button class="btn-primary" @click="keyUpdate">Save</button>
                </div>
            </div>
        </div>

        <!-- Remove Key Modal -->
        <div class="modal" :class="{ show: keyRemoveModalVisible }">
            <div class="modal-content">
                <div class="modal-header">Remove Key</div>
                <div class="alert alert-danger" v-if="keyRemoveForm.errorMessage">{{ keyRemoveForm.errorMessage }}</div>
                <p style="margin-bottom: 10px;">Are you sure you want to remove key '{{ keyRemoveForm.key }}'?</p>
                <p style="color: #dc3545; margin-bottom: 20px;">This action cannot be undone.</p>
                <div class="modal-footer">
                    <button class="btn-secondary" @click="keyRemoveModalClose">Cancel</button>
                    <button class="btn-danger" @click="keyRemove">Delete</button>
                </div>
            </div>
        </div>
    </div>
    ` + u.getJS(string(vaultPathJSON)) + `
</body>
</html>`
}
