// @ts-ignore
/* eslint-disable */
import request from '@/request'

/** 添加黑名单 POST /blacklist/add */
export async function addToBlacklist(body: API.blacklistRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: boolean }>('/blacklist/add', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}

/** 检查是否在黑名单 GET /blacklist/check */
export async function checkBlacklist(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.checkBlacklistParams,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse & { data?: boolean }>('/blacklist/check', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  })
}

/** 获取黑名单数量 GET /blacklist/count */
export async function countBlacklist(options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: number }>('/blacklist/count', {
    method: 'GET',
    ...(options || {}),
  })
}

/** 获取黑名单列表 GET /blacklist/list */
export async function getBlacklist(options?: { [key: string]: any }) {
  return request<API.BaseResponse & { data?: string[] }>('/blacklist/list', {
    method: 'GET',
    ...(options || {}),
  })
}

/** 移除黑名单 POST /blacklist/remove */
export async function removeFromBlacklist(
  body: API.blacklistRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse & { data?: boolean }>('/blacklist/remove', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  })
}
