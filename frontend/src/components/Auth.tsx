'use client';

import { useGoogleLogin, CodeResponse } from '@react-oauth/google';
import axios from 'axios';

export default function Auth() {

  const handleSuccess = async (codeResponse: Omit<CodeResponse, 'error' | 'error_description' | 'error_uri'>) => {
    console.log("Lấy được Authorization Code:", codeResponse.code);

    try {
      const response = await axios.post('http://localhost:8000/login', {
        code: codeResponse.code,
      });

      console.log("Phản hồi từ backend:", response.data);
      alert("Đăng nhập thành công! Kiểm tra console để xem chi tiết.");

    } catch (err) {
      console.error("Lỗi khi gửi code lên backend:", err);
      alert("Đăng nhập thất bại! Kiểm tra console để xem chi tiết.");
    }
  };

  const login = useGoogleLogin({
    flow: 'auth-code',
    scope: 'https://www.googleapis.com/auth/gmail.readonly',
    onSuccess: handleSuccess,
    onError: (error) => console.error("Login Failed:", error),
  });

  return (
    <button className='border border-white p-3 rounded-sm cursor-pointer' onClick={() => login()}>
      Đăng nhập với Google
    </button>
  );
}