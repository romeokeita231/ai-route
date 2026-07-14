declare namespace API {
  type BaseResponse = {
    code?: number
    data?: any
    message?: string
  }

  type blacklistRequest = {
    ip?: string
    reason?: string
  }

  type checkBlacklistParams = {
    /** IP */
    ip: string
  }

  type DeleteRequest = {
    id?: number
  }

  type LoginUserVO = {
    createTime?: string
    id?: string
    updateTime?: string
    userAccount?: string
    userAvatar?: string
    userName?: string
    userProfile?: string
    userRole?: string
  }

  type UserAddRequest = {
    userAccount?: string
    userAvatar?: string
    userName?: string
    userProfile?: string
    userRole?: string
  }

  type UserLoginRequest = {
    userAccount?: string
    userPassword?: string
  }

  type UserRegisterRequest = {
    checkPassword?: string
    userAccount?: string
    userPassword?: string
  }
}
