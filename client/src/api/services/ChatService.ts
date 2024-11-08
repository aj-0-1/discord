/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { chat_Message } from '../models/chat_Message';
import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';
export class ChatService {
    constructor(public readonly httpRequest: BaseHttpRequest) {}
    /**
     * Send message
     * Send a private message to another user
     * @param authorization Bearer token
     * @param request Message content
     * @returns chat_Message OK
     * @throws ApiError
     */
    public postChatMessages(
        authorization: string,
        request: chat_Message,
    ): CancelablePromise<chat_Message> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/chat/messages',
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
            },
        });
    }
    /**
     * Get messages
     * Get chat messages with another user
     * @param authorization Bearer token
     * @param userId User ID to get messages with
     * @returns chat_Message OK
     * @throws ApiError
     */
    public getChatMessages(
        authorization: string,
        userId: string,
    ): CancelablePromise<Array<chat_Message>> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/chat/messages/{userID}',
            path: {
                'userID': userId,
            },
            headers: {
                'Authorization': authorization,
            },
            errors: {
                401: `Unauthorized`,
            },
        });
    }
    /**
     * WebSocket connection
     * Connect to WebSocket for real-time messages
     * @param authorization Bearer token
     * @returns void
     * @throws ApiError
     */
    public getChatWs(
        authorization: string,
    ): CancelablePromise<void> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/chat/ws',
            headers: {
                'Authorization': authorization,
            },
            errors: {
                401: `Unauthorized`,
            },
        });
    }
}
