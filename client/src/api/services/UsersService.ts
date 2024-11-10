/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { user_User } from '../models/user_User';
import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';
export class UsersService {
    constructor(public readonly httpRequest: BaseHttpRequest) {}
    /**
     * Search users
     * Search users by username or email
     * @param authorization Bearer token
     * @param q Search query
     * @returns user_User OK
     * @throws ApiError
     */
    public getUsersSearch(
        authorization: string,
        q: string,
    ): CancelablePromise<Array<user_User>> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/users/search',
            headers: {
                'Authorization': authorization,
            },
            query: {
                'q': q,
            },
            errors: {
                400: `Bad request`,
                401: `Unauthorized`,
            },
        });
    }
}
