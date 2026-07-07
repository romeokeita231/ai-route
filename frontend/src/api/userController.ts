// @ts-ignore
/* eslint-disable */
import request from '@/request'

/** 添加用户 POST /user/add */
export async function addUser(body: API.UserAddRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: number }>('/user/add', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}

/** 删除用户 POST /user/delete */
export async function deleteUser(body: API.DeleteRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: boolean }>('/user/delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}

/** 获取登录用户信息 GET /user/login */
export async function getLoginUser(options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: API.LoginUserVO }>('/user/login', {
    method: 'GET',
    ...(options || {}),
  })
}

/** 用户登录 POST /user/login */
export async function userLogin(body: API.UserLoginRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: API.LoginUserVO }>('/user/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}

/** 用户退出登录 POST /user/logout */
export async function userLogout(options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: boolean }>('/user/logout', {
    method: 'POST',
    ...(options || {}),
  })
}

/** 用户注册 POST /user/register */
export async function userRegister(
  body: API.UserRegisterRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse & { data?: string }>('/user/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}
