
const API_BASE_URL = 'http://localhost:8081/api/v1';

// 检查登录状态
function checkAuth() {
    const token = localStorage.getItem('token');
    if (!token) return false;
    return token.split('.').length === 3;
}

// 路由跳转
function navigateTo(path) {
    if (path === '/friends' && !checkAuth()) {
        window.location.href = '/';
        return;
    }
    window.location.href = path;
}

// 页面加载时检查路由
if (window.location.pathname !== '/' && checkAuth()) {
    navigateTo('/friends');
}

new Vue({
    el: '#app',
    data: {
        isLoginForm: true,
        loginForm: {
            email: '',
            password: ''
        },
        registerForm: {
            email: '',
            password: '',
            confirmPassword: '',
            username: ''
        },
        errorMessage: '',
        successMessage: '',
        isLoading: false
    },
    methods: {
        toggleForm() {
            this.isLoginForm = !this.isLoginForm;
            this.errorMessage = '';
            this.successMessage = '';
        },
        validateLogin() {
            if (!this.loginForm.email || !this.loginForm.password) {
                this.errorMessage = '请填写邮箱/用户ID和密码';
                return false;
            }
            return true;
        },
        validateRegister() {
            if (!this.registerForm.email || !this.registerForm.password ||
                !this.registerForm.confirmPassword || !this.registerForm.username) {
                this.errorMessage = '请正确填写邮箱和密码';
                return false;
            }

            if (this.registerForm.password !== this.registerForm.confirmPassword) {
                this.errorMessage = '两次输入的密码不一致';
                return false;
            }

            if (this.registerForm.password.length < 6) {
                this.errorMessage = '密码长度不能少于6位';
                return false;
            }

            return true;
        },
        async handleLogin() {
            if (!this.validateLogin()) return;

            this.isLoading = true;
            this.errorMessage = '';

            try {
                const response = await axios.post(`${API_BASE_URL}/login`, {
                    identifier: this.loginForm.email || this.loginForm.uid,
                    password: this.loginForm.password
                }, {
                    headers: {
                        'Accept': 'application/json'
                    }
                });

                if (response.data && response.data.data) {
                    const username = response.data.data.username || '用户';
                    alert(`欢迎用户 ${username}`);
                    // 保存token到localStorage
                    localStorage.setItem('token', response.data.data.token);
                    // 立即跳转到好友界面
                    navigateTo('/friends');
                } else {
                    this.errorMessage = response.data.message || '登录失败，请重试';
                }
            } catch (error) {
                if (error.response) {
                    switch (error.response.status) {
                        case 400:
                            this.errorMessage = error.response.data.msg || '参数格式错误';
                            break;
                        case 401:
                            this.errorMessage = '密码错误';
                            break;
                        case 404:
                            this.errorMessage = '用户不存在';
                            break;
                        case 500:
                            this.errorMessage = '服务器内部错误';
                            break;
                        default:
                            this.errorMessage = '登录失败';
                    }
                } else {
                    this.errorMessage = '网络错误，请稍后重试';
                }
            } finally {
                this.isLoading = false;
            }
        },
        async handleRegister() {
            if (!this.validateRegister()) return;

            this.isLoading = true;
            this.errorMessage = '';

            try {
                const response = await axios.post(`${API_BASE_URL}/register`, {
                    email: this.registerForm.email,
                    password: this.registerForm.password,
                    rePassword: this.registerForm.confirmPassword,
                    username: this.registerForm.username
                });

                if (response.data.code === 200) {
                    this.successMessage = `注册成功！您的用户ID是: ${response.data.data}`;

                    // 使用 $nextTick 确保DOM更新
                    this.$nextTick(() => {
                        alert(this.successMessage);
                        // 自动切换到登录表单
                        setTimeout(() => {
                            this.isLoginForm = true;
                            this.successMessage = '';
                        }, 3000);
                    });
                } else {
                    // 优先显示后端返回的具体错误信息
                    this.errorMessage = response.data.msg ||
                                      (response.data.errors ? response.data.errors[0] : '注册失败');
                }
            } catch (error) {
                if (error.response) {
                    switch (error.response.status) {
                        case 400:
                            this.errorMessage = error.response.data.msg || '参数格式错误';
                            break;
                        case 422:
                            this.errorMessage = '该邮箱已被注册';
                            break;
                        case 500:
                            this.errorMessage = '服务器内部错误';
                            break;
                        default:
                            this.errorMessage = '注册失败';
                    }
                } else {
                    this.errorMessage = '网络错误，请稍后重试';
                }
            } finally {
                this.isLoading = false;
            }
        }
    },
    template: `
        <div id="auth-container">
            <h1 class="welcome-title">欢迎进入社交平台系统</h1>
            <div v-if="isLoginForm" class="auth-form">
                <h2 class="form-title">用户登录</h2>
                <form @submit.prevent="handleLogin">
                    <div class="form-group">
                        <label for="login-email">邮箱/用户ID</label>
                        <input 
                            id="login-email" 
                            type="text" 
                            class="form-control" 
                            v-model="loginForm.email" 
                            placeholder="请输入邮箱或用户ID"
                        >
                    </div>
                    <div class="form-group">
                        <label for="login-password">密码</label>
                        <input 
                            id="login-password" 
                            type="password" 
                            class="form-control" 
                            v-model="loginForm.password" 
                            placeholder="请输入密码"
                        >
                    </div>
                    <button type="submit" class="btn btn-primary" :disabled="isLoading">
                        {{ isLoading ? '登录中...' : '登录' }}
                    </button>
                    <div class="form-footer">
                        <p>还没有账号？<button type="button" class="toggle-btn" @click="toggleForm">注册账号</button></p>
                    </div>
                </form>
            </div>

            <div v-else class="auth-form">
                <h2 class="form-title">用户注册</h2>
                <form @submit.prevent="handleRegister">
                    <div class="form-group">
                        <label for="register-email">邮箱</label>
                        <input 
                            id="register-email" 
                            type="email" 
                            class="form-control" 
                            v-model="registerForm.email" 
                            placeholder="请输入邮箱"
                        >
                    </div>
                    <div class="form-group">
                        <label for="register-username">用户名</label>
                        <input 
                            id="register-username" 
                            type="text" 
                            class="form-control" 
                            v-model="registerForm.username" 
                            placeholder="请输入用户名"
                        >
                    </div>
                    <div class="form-group">
                        <label for="register-password">密码</label>
                        <input 
                            id="register-password" 
                            type="password" 
                            class="form-control" 
                            v-model="registerForm.password" 
                            placeholder="请输入密码"
                        >
                    </div>
                    <div class="form-group">
                        <label for="register-confirm-password">确认密码</label>
                        <input 
                            id="register-confirm-password" 
                            type="password" 
                            class="form-control" 
                            v-model="registerForm.confirmPassword" 
                            placeholder="请再次输入密码"
                        >
                    </div>
                    <button type="submit" class="btn btn-primary" :disabled="isLoading">
                        {{ isLoading ? '注册中...' : '注册' }}
                    </button>
                    <div class="form-footer">
                        <p>已有账号？<button type="button" class="toggle-btn" @click="toggleForm">登录账号</button></p>
                    </div>
                </form>
            </div>

            <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
            <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
        </div>
        <style>
            .welcome-title {
                text-align: center;
                color: #333;
                margin-bottom: 20px;
                font-size: 24px;
            }
        </style>
    `
});
