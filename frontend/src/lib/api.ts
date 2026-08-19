import axios from 'axios';
import Cookies from 'js-cookie';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export const api = axios.create({
  baseURL: API_URL,
  withCredentials: true, // For sending cookies
});

// Add a request interceptor to attach the token if available in cookies or localStorage
api.interceptors.request.use((config) => {
  let token = Cookies.get('token');
  if (typeof window !== 'undefined' && !token) {
    token = localStorage.getItem('token');
  }
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Add a response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If the error status is 401 and there is no originalRequest._retry flag
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = Cookies.get('refresh_token');
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const response = await axios.post(`${API_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        if (response.data.success) {
          const newToken = response.data.token;
          const newRefreshToken = response.data.refresh_token;

          // Save the new tokens
          Cookies.set('token', newToken, { secure: true, sameSite: 'strict' });
          if (newRefreshToken) {
            Cookies.set('refresh_token', newRefreshToken, { secure: true, sameSite: 'strict' });
          }

          // Retry the original request
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return api(originalRequest);
        }
      } catch (refreshError) {
        // If refresh fails, redirect to login
        Cookies.remove('token');
        Cookies.remove('refresh_token');
        
        // Use window location to redirect to avoid circular dependencies with next/router
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
