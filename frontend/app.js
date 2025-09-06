const API_BASE = 'http://localhost:8080/api/v1';
let currentUser = null;
let authToken = localStorage.getItem('authToken');
let captchaToken = null;

// Turnstile callback
window.onTurnstileSuccess = function(token) {
    captchaToken = token;
};

// DOM элементы
const loginForm = document.getElementById('loginForm');
const signupForm = document.getElementById('signupForm');
const verifyForm = document.getElementById('verifyForm');
const profileSection = document.getElementById('profileSection');
const editProfileSection = document.getElementById('editProfileSection');
const userInfo = document.getElementById('userInfo');
const logoutBtn = document.getElementById('logoutBtn');

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    if (authToken) {
        loadProfile();
    }
    setupEventListeners();
});

// События
function setupEventListeners() {
    // Переключение форм
    document.getElementById('showSignup').onclick = (e) => {
        e.preventDefault();
        showSignup();
    };
    
    document.getElementById('showLogin').onclick = (e) => {
        e.preventDefault();
        showLogin();
    };
    
    document.getElementById('showVerify').onclick = (e) => {
        e.preventDefault();
        showVerify();
    };
    
    document.getElementById('backToLogin').onclick = (e) => {
        e.preventDefault();
        showLogin();
    };

    // Формы
    document.getElementById('loginFormElement').onsubmit = handleLogin;
    document.getElementById('signupFormElement').onsubmit = handleSignup;
    document.getElementById('verifyFormElement').onsubmit = handleVerify;
    document.getElementById('profileForm').onsubmit = handleProfileUpdate;
    
    // Аватар
    document.getElementById('uploadAvatarBtn').onclick = handleAvatarUpload;
    
    // Выход
    logoutBtn.onclick = handleLogout;
    
    // Редактирование профиля
    document.getElementById('editProfileBtn').onclick = showEditProfile;
    document.getElementById('cancelEditBtn').onclick = showProfile;
    document.getElementById('cancelEditBtn2').onclick = showProfile;
}

// API запросы
async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const config = {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        },
        ...options
    };
    
    if (authToken && !config.headers.Authorization) {
        config.headers.Authorization = `Bearer ${authToken}`;
    }
    
    try {
        const response = await fetch(url, config);
        
        let data;
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            const text = await response.text();
            data = { message: text || 'Пустой ответ' };
        }
        
        if (!response.ok) {
            throw new Error(data.error?.message || data.message || 'Ошибка запроса');
        }
        
        return data;
    } catch (error) {
        showAlert(error.message, 'danger');
        throw error;
    }
}

// Обработчики форм
async function handleLogin(e) {
    e.preventDefault();
    
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    
    try {
        const data = await apiRequest('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        
        authToken = data.access_token;
        localStorage.setItem('authToken', authToken);
        currentUser = data.user;
        
        showProfile();
        showAlert('Вход выполнен успешно!', 'success');
    } catch (error) {
        console.error('Login error:', error);
    }
}

async function handleSignup(e) {
    e.preventDefault();
    
    const email = document.getElementById('signupEmail').value;
    const nick = document.getElementById('signupNick').value;
    const password = document.getElementById('signupPassword').value;
    const confirm_password = document.getElementById('signupConfirmPassword').value;
    const agree_terms = document.getElementById('agreeTerms').checked;
    
    if (!captchaToken) {
        showAlert('Пройдите проверку капчи', 'warning');
        return;
    }
    
    try {
        await apiRequest('/auth/signup', {
            method: 'POST',
            body: JSON.stringify({ email, nick, password, confirm_password, agree_terms, captcha_token: captchaToken })
        });
        
        showAlert('Регистрация успешна! Подтверди email.', 'success');
        showVerify();
    } catch (error) {
        console.error('Signup error:', error);
    }
}

async function handleVerify(e) {
    e.preventDefault();
    
    const token = document.getElementById('verifyToken').value;
    
    try {
        await apiRequest(`/auth/verify-email?token=${token}`, {
            method: 'POST'
        });
        
        showAlert('Email подтвержден! Можно войти.', 'success');
        showLogin();
    } catch (error) {
        console.error('Verify error:', error);
    }
}

async function handleProfileUpdate(e) {
    e.preventDefault();
    
    const nick = document.getElementById('profileNick').value;
    const bio = document.getElementById('profileBio').value;
    
    try {
        await apiRequest('/profile', {
            method: 'PUT',
            body: JSON.stringify({ nick, bio })
        });
        
        showAlert('Профиль обновлен!', 'success');
        loadProfile();
        showProfile();
    } catch (error) {
        console.error('Profile update error:', error);
    }
}

async function handleAvatarUpload() {
    const fileInput = document.getElementById('avatarInput');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('Выберите файл', 'warning');
        return;
    }
    
    const formData = new FormData();
    formData.append('avatar', file);
    
    try {
        const response = await fetch(`${API_BASE}/profile/avatar`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`
            },
            body: formData
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error?.message || 'Ошибка загрузки');
        }
        
        // Обновляем аватар с timestamp чтобы обойти кеш
        const avatarUrl = data.avatar_url + '?t=' + Date.now();
        document.getElementById('avatarPreview').src = avatarUrl;
        document.getElementById('editAvatarPreview').src = avatarUrl;
        showAlert('Аватар загружен!', 'success');
    } catch (error) {
        showAlert(error.message, 'danger');
    }
}

function handleLogout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    showLogin();
    showAlert('Вы вышли из системы', 'info');
}

// Загрузка данных
async function loadProfile() {
    try {
        const user = await apiRequest('/profile');
        currentUser = user;
        
        // Обновляем просмотр профиля
        document.getElementById('profileDisplayName').textContent = user.nick || 'Пользователь';
        document.getElementById('profileEmail').textContent = user.email || '';
        document.getElementById('profileBioDisplay').textContent = user.bio || 'Информация не указана';
        
        // Обновляем форму редактирования
        document.getElementById('profileNick').value = user.nick || '';
        document.getElementById('profileBio').value = user.bio || '';
        document.getElementById('profileEmailInput').value = user.email || '';
        
        if (user.avatar) {
            const avatarUrl = user.avatar + '?t=' + Date.now();
            document.getElementById('avatarPreview').src = avatarUrl;
            document.getElementById('editAvatarPreview').src = avatarUrl;
        }
        
        // Загружаем сабмиты
        loadSubmissions();
        
        showProfile();
    } catch (error) {
        console.error('Load profile error:', error);
        handleLogout();
    }
}

async function loadSubmissions() {
    try {
        const data = await apiRequest('/profile/submissions');
        const container = document.getElementById('submissionsList');
        
        if (data.submissions.length === 0) {
            container.innerHTML = '<p class="text-muted">Нет сабмитов</p>';
            return;
        }
        
        const count = data.submissions.length;
        document.getElementById('submissionsCount').textContent = count;
        document.getElementById('submissionsBadge').textContent = count;
        
        container.innerHTML = data.submissions.map(sub => `
            <div class="card submission-card">
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-start">
                        <div>
                            <h6 class="mb-1 text-primary">${sub.game_title}</h6>
                            <p class="mb-1 text-muted">${sub.jam_title}</p>
                            <small class="text-success">✓ Отправлено</small>
                        </div>
                        <span class="badge bg-light text-dark">${new Date(sub.submitted_at).toLocaleDateString()}</span>
                    </div>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Load submissions error:', error);
    }
}

// UI функции
function showLogin() {
    loginForm.classList.remove('hidden');
    signupForm.classList.add('hidden');
    verifyForm.classList.add('hidden');
    profileSection.classList.add('hidden');
    userInfo.classList.add('hidden');
    logoutBtn.classList.add('hidden');
}

function showSignup() {
    loginForm.classList.add('hidden');
    signupForm.classList.remove('hidden');
    verifyForm.classList.add('hidden');
    profileSection.classList.add('hidden');
    userInfo.classList.add('hidden');
    logoutBtn.classList.add('hidden');
}

function showVerify() {
    loginForm.classList.add('hidden');
    signupForm.classList.add('hidden');
    verifyForm.classList.remove('hidden');
    profileSection.classList.add('hidden');
    userInfo.classList.add('hidden');
    logoutBtn.classList.add('hidden');
}

function showProfile() {
    loginForm.classList.add('hidden');
    signupForm.classList.add('hidden');
    verifyForm.classList.add('hidden');
    profileSection.classList.remove('hidden');
    editProfileSection.classList.add('hidden');
    userInfo.classList.remove('hidden');
    logoutBtn.classList.remove('hidden');
    
    if (currentUser) {
        userInfo.textContent = `Привет, ${currentUser.nick}!`;
    }
}

function showEditProfile() {
    loginForm.classList.add('hidden');
    signupForm.classList.add('hidden');
    verifyForm.classList.add('hidden');
    profileSection.classList.add('hidden');
    editProfileSection.classList.remove('hidden');
    userInfo.classList.remove('hidden');
    logoutBtn.classList.remove('hidden');
}

function showAlert(message, type = 'info') {
    const alertsContainer = document.getElementById('alerts');
    const alert = document.createElement('div');
    alert.className = `alert alert-${type} alert-dismissible fade show position-fixed top-0 end-0 m-3`;
    alert.style.zIndex = '9999';
    alert.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    alertsContainer.appendChild(alert);
    
    setTimeout(() => {
        alert.remove();
    }, 5000);
}