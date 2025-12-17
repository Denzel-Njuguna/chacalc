1 Signup

Request

curl -X POST http://localhost:8080/auth/signup   -H "Content-Type: application/json"   -d '{"email":"test@gmail.com","password":"password123"}'

Response (success)

{
"message": "Signup successful."
}

2 Login

Request

curl -X POST http://localhost:8080/auth/login   -H "Content-Type: application/json"   -d '{"email":"test@gmail.com","password":"password123"}'


Response (success)

{
"access_token": "<ACCESS_TOKEN>",
"token_type": "bearer",
"expires_in": 3600,
"refresh_token": "<REFRESH_TOKEN>",
"user": { ... }
}


Copy access_token for protected routes.

3 Logout

Request

curl -X POST http://localhost:8080/auth/logout   -H "Authorization: Bearer <access token>"



Response (success)

HTTP 204 No Content


Use the access token. This logs the user out and invalidates the session.

4 Forgot Password

Request

curl -X POST http://localhost:8080/auth/forgot-password \ -H "Content-Type: application/json" \ -d '{"email":"test@example.com"}'

Response (success)

{
"message": "If an account exists, a reset email has been sent"
}


Supabase sends a recovery link to the email.

5 Reset Password

Request

curl -X POST http://localhost:8080/auth/reset-password \
-H "Authorization: Bearer <RECOVERY_ACCESS_TOKEN>" \
-H "Content-Type: application/json" \
-d '{
"password":"NewStrongPassword123"
}'


Response (success)

{
"message": "Password reset successful. Please log in again."
}


Use the recovery access token sent by Supabase.

6 Update Password (Authenticated)

Request

curl -X POST http://localhost:8080/auth/update-password \
-H "Authorization: Bearer <ACCESS_TOKEN>" \
-H "Content-Type: application/json" \
-d '{
"new_password": "NewPassword456"
}'


Response (success)

{
"message": "Password updated successfully. Please log in again."
}


Use the current JWT. After update, the old token is invalidated.

7 Get Current User (/me)

Request

curl -X GET http://localhost:8080/auth/me \
-H "Authorization: Bearer <ACCESS_TOKEN>"


Response (success)

{
"id": "user-id",
"email": "test@example.com",
"role": "authenticated",
"app_metadata": { ... },
"user_metadata": { ... }
}

8 Onboard User

request

curl -X POST http://localhost:8080/auth/onboarding \
-H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
-H "Content-Type: application/json" \
-d '{
"username": "Test",
"phone": "0123456789"
}'
