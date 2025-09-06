# Frontend Helper - Hubigr API

## 🚀 Быстрый старт

### 1. Настройка
```bash
# .env
VITE_API_URL=http://localhost:8000/api/v1
```

### 2. Базовый запрос
```javascript
const API_BASE = import.meta.env.VITE_API_URL;

const response = await fetch(`${API_BASE}/health`);
const data = await response.json();
```

---

## 🔐 Аутентификация

### Регистрация
```javascript
const register = async (userData) => {
  const response = await fetch(`${API_BASE}/auth/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: userData.email,
      password: userData.password,
      confirm_password: userData.password,
      nick: userData.nick,
      agree_terms: true
    })
  });

  if (response.ok) {
    const data = await response.json();
    console.log('Регистрация успешна:', data);
    return data;
  } else {
    const error = await response.json();
    throw new Error(error.error.message);
  }
};
```

### Вход
```javascript
const login = async (email, password) => {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (response.ok) {
    const { user, token } = await response.json();
    localStorage.setItem('token', token);
    return { user, token };
  } else {
    const error = await response.json();
    throw new Error(error.error.message);
  }
};
```

### Запросы с токеном
```javascript
const apiRequest = async (endpoint, options = {}) => {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers
    },
    ...options
  });

  return response;
};
```

---

## 👤 Работа с профилем

### Получить профиль
```javascript
const getProfile = async () => {
  const response = await apiRequest('/profile');

  if (response.ok) {
    return await response.json();
  } else {
    throw new Error('Failed to get profile');
  }
};
```

### Обновить профиль
```javascript
const updateProfile = async (profileData) => {
  const response = await apiRequest('/profile', {
    method: 'PUT',
    body: JSON.stringify(profileData)
  });

  if (response.ok) {
    return await response.json();
  } else {
    const error = await response.json();
    throw new Error(error.error.message);
  }
};
```

### Загрузить аватар
```javascript
const uploadAvatar = async (file) => {
  const formData = new FormData();
  formData.append('avatar', file);

  const response = await fetch(`${API_BASE}/profile/avatar`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('token')}`
    },
    body: formData
  });

  if (response.ok) {
    const data = await response.json();
    return data.avatar_url;
  } else {
    throw new Error('Upload failed');
  }
};
```

---

## ⚠️ Обработка ошибок

### Стандартные коды ошибок
```javascript
const handleApiError = (response) => {
  switch (response.status) {
    case 400:
      return 'Неверные данные';
    case 401:
      localStorage.removeItem('token');
      window.location.href = '/login';
      return 'Не авторизован';
    case 403:
      return 'Доступ запрещен';
    case 404:
      return 'Не найдено';
    case 422:
      return 'Ошибка валидации';
    case 429:
      return 'Слишком много запросов';
    case 500:
      return 'Ошибка сервера';
    default:
      return 'Неизвестная ошибка';
  }
};
```

### Использование
```javascript
try {
  const data = await getProfile();
  console.log('Профиль:', data);
} catch (error) {
  const message = handleApiError(error.response);
  alert(message);
}
```

---

## 📋 Формы валидации

### Регистрация
```javascript
const validateRegistration = (data) => {
  const errors = [];

  if (!data.email.includes('@')) {
    errors.push('Неверный email');
  }

  if (data.password.length < 6) {
    errors.push('Пароль минимум 6 символов');
  }

  if (data.nick.length < 2 || data.nick.length > 50) {
    errors.push('Ник 2-50 символов');
  }

  return errors;
};
```

### Профиль
```javascript
const validateProfile = (data) => {
  const errors = [];

  if (data.bio && data.bio.length > 200) {
    errors.push('Био максимум 200 символов');
  }

  if (data.links && data.links.length > 5) {
    errors.push('Максимум 5 ссылок');
  }

  return errors;
};
```

---

## 🔄 Авторизация в компонентах

### React Hook
```javascript
import { useState, useEffect } from 'react';

const useAuth = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const profile = await getProfile();
        setUser(profile);
      } catch (error) {
        localStorage.removeItem('token');
      }
    }
    setLoading(false);
  };

  const logout = () => {
    localStorage.removeItem('token');
    setUser(null);
  };

  return { user, loading, logout, checkAuth };
};
```

### Защищенные маршруты
```javascript
const PrivateRoute = ({ children }) => {
  const { user, loading } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (!user) return <Navigate to="/login" />;

  return children;
};
```

---

## 📊 Типичные сценарии

### 1. Регистрация → Вход → Профиль
```javascript
// 1. Регистрация
await register(userData);

// 2. Вход
const { user, token } = await login(email, password);

// 3. Получение профиля
const profile = await getProfile();

// 4. Обновление профиля
await updateProfile({ nick: 'NewNick' });
```

### 2. Загрузка аватара
```javascript
const handleAvatarUpload = async (file) => {
  try {
    const avatarUrl = await uploadAvatar(file);
    await updateProfile({ avatar: avatarUrl });
    alert('Аватар обновлен!');
  } catch (error) {
    alert('Ошибка загрузки');
  }
};
```

### 3. Обработка 401 ошибки
```javascript
const response = await apiRequest('/profile');

if (response.status === 401) {
  localStorage.removeItem('token');
  window.location.href = '/login';
  return;
}
```

---

## 🎯 Полезные советы

### 1. Всегда проверяй токен
```javascript
const token = localStorage.getItem('token');
if (!token) {
  window.location.href = '/login';
  return;
}
```

### 2. Обрабатывай все ошибки
```javascript
try {
  const data = await apiRequest('/endpoint');
} catch (error) {
  console.error('API Error:', error);
  // Покажи пользователю понятное сообщение
}
```

### 3. Используй loading состояния
```javascript
const [loading, setLoading] = useState(false);

const handleSubmit = async () => {
  setLoading(true);
  try {
    await apiRequest('/endpoint', { method: 'POST' });
  } finally {
    setLoading(false);
  }
};
```

---

Если что-то не работает:
1. Проверь логи сервера: `docker-compose logs -f auth`
2. Проверь API: `curl http://localhost:8000/api/v1/health`
3. Проверь токен: `console.log(localStorage.getItem('token'))`
