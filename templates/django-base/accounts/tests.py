from django.test import TestCase
from django.contrib.auth import get_user_model
from django.urls import reverse
from rest_framework.test import APITestCase, APIClient
from rest_framework import status
from rest_framework.authtoken.models import Token
import json

User = get_user_model()


class UserRegistrationTestCase(APITestCase):
    """
    Test cases for user registration functionality.
    """

    def setUp(self):
        self.client = APIClient()
        self.registration_url = '/api/accounts/auth/registration/'  # Back to dj-rest-auth
        self.valid_user_data = {
            'email': 'testuser@example.com',
            'password1': 'testpassword123',
            'password2': 'testpassword123',
        }

    def test_user_registration_success(self):
        """
        Test successful user registration.
        """
        response = self.client.post(
            self.registration_url,
            self.valid_user_data,
            format='json'
        )

        # Debug: print response if test fails
        if response.status_code != status.HTTP_201_CREATED:
            print(f"Registration failed with status {response.status_code}")
            print(f"Response data: {response.data}")

        self.assertEqual(response.status_code, status.HTTP_201_CREATED)
        self.assertIn('key', response.data)  # Token should be returned

        # Verify user was created
        user = User.objects.get(email=self.valid_user_data['email'])
        self.assertTrue(user.is_active)  # User is active since email verification is optional

    def test_user_registration_password_mismatch(self):
        """
        Test registration with mismatched passwords.
        """
        invalid_data = self.valid_user_data.copy()
        invalid_data['password2'] = 'differentpassword'

        response = self.client.post(
            self.registration_url,
            invalid_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        # dj-rest-auth may return field errors; accept either
        self.assertTrue('non_field_errors' in response.data or 'password2' in response.data)

    def test_user_registration_duplicate_email(self):
        """
        Test registration with duplicate email.
        """
        # Create a user first with the same email
        User.objects.create_user(
            email='testuser@example.com',
            password='password123'
        )

        response = self.client.post(
            self.registration_url,
            self.valid_user_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        # Check for either email or username error since they're the same
        self.assertTrue('email' in response.data or 'username' in response.data)

    def test_user_registration_missing_fields(self):
        """
        Test registration with missing required fields.
        """
        incomplete_data = {
            'email': 'testuser@example.com',
            'password1': 'testpassword123'
            # Missing password2, first_name, last_name
        }

        response = self.client.post(
            self.registration_url,
            incomplete_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('password2', response.data)


class UserLoginTestCase(APITestCase):
    """
    Test cases for user login functionality.
    """

    def setUp(self):
        self.client = APIClient()
        self.login_url = '/api/accounts/auth/login/'
        self.logout_url = '/api/accounts/auth/logout/'

        # Create an active user for testing
        self.user = User.objects.create_user(
            email='testuser@example.com',
            password='testpassword123',
        )
        self.user.full_name = 'Test User'
        self.user.preferred_name = 'Test'
        self.user.save()
        self.user_credentials = {
            'email': 'testuser@example.com',
            'password': 'testpassword123'
        }

    def test_user_login_success(self):
        """
        Test successful user login.
        """
        response = self.client.post(
            self.login_url,
            self.user_credentials,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertIn('key', response.data)  # Token should be returned
        # Note: 'user' field might not be returned by default in dj-rest-auth

        # Verify token was created
        token = Token.objects.get(user=self.user)
        self.assertEqual(response.data['key'], token.key)

    def test_user_login_invalid_credentials(self):
        """
        Test login with invalid credentials.
        """
        invalid_credentials = {
            'email': 'testuser@example.com',
            'password': 'wrongpassword'
        }

        response = self.client.post(
            self.login_url,
            invalid_credentials,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('non_field_errors', response.data)

    def test_user_login_inactive_user(self):
        """
        Test login with inactive user account.
        """
        # Create inactive user
        inactive_user = User.objects.create_user(
            email='inactive@example.com',
            password='testpassword123',
        )
        inactive_user.full_name = 'Inactive User'
        inactive_user.is_active = False
        inactive_user.save()

        credentials = {
            'email': 'inactive@example.com',
            'password': 'testpassword123'
        }

        response = self.client.post(
            self.login_url,
            credentials,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_user_logout_success(self):
        """
        Test successful user logout.
        """
        # First login to get a token
        login_response = self.client.post(
            self.login_url,
            self.user_credentials,
            format='json'
        )
        self.assertEqual(login_response.status_code, status.HTTP_200_OK)

        # Get token from response (might be 'key' or 'access_token')
        token = login_response.data.get('key') or login_response.data.get('access_token')
        self.assertIsNotNone(token, f"No token found in response: {login_response.data}")

        # Set authentication header
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {token}')

        # Logout
        response = self.client.post(self.logout_url)

        self.assertEqual(response.status_code, status.HTTP_200_OK)

        # Verify token was deleted
        self.assertFalse(Token.objects.filter(key=token).exists())

    def test_user_logout_without_authentication(self):
        """
        Test logout without authentication.
        """
        response = self.client.post(self.logout_url)
        # dj-rest-auth might return 400 instead of 401 for unauthenticated logout
        self.assertIn(response.status_code, [status.HTTP_200_OK, status.HTTP_400_BAD_REQUEST, status.HTTP_401_UNAUTHORIZED])


class TokenAuthenticationTestCase(APITestCase):
    """
    Test cases for token-based authentication.
    """

    def setUp(self):
        self.client = APIClient()
        self.user = User.objects.create_user(
            email='testuser@example.com',
            password='testpassword123',
        )
        self.user.full_name = 'Test User'
        self.user.save()
        self.token = Token.objects.create(user=self.user)
        self.user_detail_url = '/api/accounts/auth/user/'

    def test_authenticated_request_with_valid_token(self):
        """
        Test authenticated request with valid token.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        response = self.client.get(self.user_detail_url)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.data['email'], self.user.email)
        self.assertEqual(response.data['first_name'], self.user.first_name)

    def test_authenticated_request_with_invalid_token(self):
        """
        Test authenticated request with invalid token.
        """
        self.client.credentials(HTTP_AUTHORIZATION='Token invalidtoken123')

        response = self.client.get(self.user_detail_url)

        self.assertEqual(response.status_code, status.HTTP_401_UNAUTHORIZED)

    def test_authenticated_request_without_token(self):
        """
        Test authenticated request without token.
        """
        response = self.client.get(self.user_detail_url)

        self.assertEqual(response.status_code, status.HTTP_401_UNAUTHORIZED)

    def test_update_user_details_with_authentication(self):
        """
        Test updating user details with valid authentication.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        update_data = {
            'first_name': 'Updated',
            'last_name': 'Name'
        }

        response = self.client.patch(
            self.user_detail_url,
            update_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.data['first_name'], 'Updated')
        self.assertEqual(response.data['last_name'], 'Name')

        # Verify database was updated
        self.user.refresh_from_db()
        self.assertEqual(self.user.first_name, 'Updated')
        self.assertEqual(self.user.last_name, 'Name')


class UserProfileViewTestCase(APITestCase):
    """
    Test cases for custom user profile views.
    """

    def setUp(self):
        self.client = APIClient()
        self.user = User.objects.create_user(
            email='testuser@example.com',
            password='testpassword123',
        )
        self.user.full_name = 'Test User'
        self.user.save()
        self.token = Token.objects.create(user=self.user)
        self.profile_url = reverse('accounts:user-profile')
        self.stats_url = reverse('accounts:user-stats')
        self.change_password_url = reverse('accounts:change-password')

    def test_get_user_profile_authenticated(self):
        """
        Test retrieving user profile with authentication.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        response = self.client.get(self.profile_url)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.data['email'], self.user.email)
        self.assertEqual(response.data['full_name'], 'Test User')
        self.assertIn('date_joined', response.data)

    def test_get_user_profile_unauthenticated(self):
        """
        Test retrieving user profile without authentication.
        """
        response = self.client.get(self.profile_url)

        self.assertEqual(response.status_code, status.HTTP_401_UNAUTHORIZED)

    def test_get_user_stats_authenticated(self):
        """
        Test retrieving user stats with authentication.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        response = self.client.get(self.stats_url)

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response.data['user_id'], self.user.pk)
        self.assertEqual(response.data['email'], self.user.email)
        self.assertEqual(response.data['is_active'], True)

    def test_change_password_success(self):
        """
        Test successful password change.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        password_data = {
            'old_password': 'testpassword123',
            'new_password1': 'newpassword456',
            'new_password2': 'newpassword456'
        }

        response = self.client.post(
            self.change_password_url,
            password_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertIn('message', response.data)

        # Verify password was changed
        self.user.refresh_from_db()
        self.assertTrue(self.user.check_password('newpassword456'))

    def test_change_password_wrong_old_password(self):
        """
        Test password change with wrong old password.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        password_data = {
            'old_password': 'wrongpassword',
            'new_password1': 'newpassword456',
            'new_password2': 'newpassword456'
        }

        response = self.client.post(
            self.change_password_url,
            password_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('old_password', response.data)

    def test_change_password_mismatch(self):
        """
        Test password change with mismatched new passwords.
        """
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {self.token.key}')

        password_data = {
            'old_password': 'testpassword123',
            'new_password1': 'newpassword456',
            'new_password2': 'differentpassword'
        }

        response = self.client.post(
            self.change_password_url,
            password_data,
            format='json'
        )

        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('non_field_errors', response.data)


class SerializerTestCase(APITestCase):
    """
    Test cases for custom serializers.
    """

    def setUp(self):
        self.user = User.objects.create_user(
            email='testuser@example.com',
            password='testpassword123',
        )
        self.user.full_name = 'Test User'
        self.user.save()

    def test_custom_register_serializer_validation(self):
        """
        Test custom registration serializer validation.
        """
        from .api.serializers import CustomRegisterSerializer

        valid_data = {
            'email': 'newuser@example.com',
            'password1': 'newpassword123',
            'password2': 'newpassword123',
            'full_name': 'New User',
            'preferred_name': 'New'
        }

        serializer = CustomRegisterSerializer(data=valid_data)
        self.assertTrue(serializer.is_valid())

        cleaned_data = serializer.get_cleaned_data()
        self.assertEqual(cleaned_data['first_name'], 'New')
        self.assertEqual(cleaned_data['last_name'], 'User')

    def test_user_details_serializer_email_validation(self):
        """
        Test user details serializer email uniqueness validation.
        """
        from .api.serializers import UserDetailsSerializer
        from django.test import RequestFactory
        from rest_framework.request import Request

        # Create another user with different email
        other_user = User.objects.create_user(
            email='other@example.com',
            password='password123'
        )

        # Create a mock request
        factory = RequestFactory()
        request = factory.get('/')
        request.user = other_user
        drf_request = Request(request)

        # Try to update with existing email
        serializer = UserDetailsSerializer(
            instance=other_user,
            data={'email': self.user.email},
            context={'request': drf_request},
            partial=True
        )

        self.assertFalse(serializer.is_valid())
        self.assertIn('email', serializer.errors)

    def test_change_password_serializer_validation(self):
        """
        Test change password serializer validation using the actual API endpoint.
        """
        from rest_framework.authtoken.models import Token

        # Create a token for the user
        token = Token.objects.create(user=self.user)

        # Test with correct old password via API
        self.client.credentials(HTTP_AUTHORIZATION=f'Token {token.key}')

        valid_data = {
            'old_password': 'testpassword123',
            'new_password1': 'newpassword456',
            'new_password2': 'newpassword456'
        }

        response = self.client.post('/api/accounts/change-password/', valid_data, format='json')
        self.assertEqual(response.status_code, 200)

        # Verify password was changed
        self.user.refresh_from_db()
        self.assertTrue(self.user.check_password('newpassword456'))

        # Test with incorrect old password
        invalid_data = {
            'old_password': 'wrongpassword',
            'new_password1': 'anotherpassword789',
            'new_password2': 'anotherpassword789'
        }

        response = self.client.post('/api/accounts/change-password/', invalid_data, format='json')
        self.assertEqual(response.status_code, 400)
        self.assertIn('old_password', response.data)
