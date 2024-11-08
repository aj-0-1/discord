/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { auth_AuthResponse } from '../models/auth_AuthResponse';
import type { auth_LoginRequest } from '../models/auth_LoginRequest';
import type { auth_RegisterRequest } from '../models/auth_RegisterRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';
export class AuthService {
    constructor(public readonly httpRequest: BaseHttpRequest) {}
    /**
     * Login user
     * Authenticate user with email and password
     * @param request Login credentials
     * @returns auth_AuthResponse OK
     * @throws ApiError
     */
    public postAuthLogin(
        request: auth_LoginRequest,
    ): CancelablePromise<auth_AuthResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/auth/login',
            body: request,
            errors: {
                400: `Invalid request`,
                401: `Invalid credentials`,
            },
        });
    }
    /**
     * Register new user
     * Register a new user with email, password and username
     * @param request Registration credentials
     * @returns auth_AuthResponse OK
     * @throws ApiError
     */
    public postAuthRegister(
        request: auth_RegisterRequest,
    ): CancelablePromise<auth_AuthResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/auth/register',
            body: request,
            errors: {
                400: `Invalid request`,
                409: `User already exists`,
            },
        });
    }
}
