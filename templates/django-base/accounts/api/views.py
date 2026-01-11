from rest_framework import status, permissions
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from rest_framework.views import APIView
from django.contrib.auth import get_user_model
from rest_framework.authtoken.models import Token

from .serializers import UserProfileSerializer, ChangePasswordSerializer

User = get_user_model()


class UserProfileView(APIView):
    """
    Retrieve user profile information.
    """
    permission_classes = [permissions.IsAuthenticated]

    def get(self, request):
        """
        Return the current user's profile.
        """
        serializer = UserProfileSerializer(request.user)
        return Response(serializer.data)


class ChangePasswordView(APIView):
    """
    Change user password.
    """
    permission_classes = [permissions.IsAuthenticated]

    def post(self, request):
        """
        Change the user's password.
        """
        serializer = ChangePasswordSerializer(
            data=request.data,
            context={'request': request}
        )
        if serializer.is_valid():
            serializer.save()
            return Response(
                {'message': 'Password changed successfully.'},
                status=status.HTTP_200_OK
            )
        return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)
@api_view(['POST'])
@permission_classes([permissions.AllowAny])
def simple_register(request):
    email = request.data.get('email')
    password1 = request.data.get('password1')
    password2 = request.data.get('password2')

    errors = {}
    if not email:
        errors['email'] = ['This field is required.']
    if not password1:
        errors['password1'] = ['This field is required.']
    if not password2:
        errors['password2'] = ['This field is required.']
    if errors:
        return Response(errors, status=status.HTTP_400_BAD_REQUEST)

    if password1 != password2:
        return Response({'password2': ['The two password fields didn\'t match.']}, status=status.HTTP_400_BAD_REQUEST)

    if User.objects.filter(email=email).exists():
        return Response({'email': ['A user with this email already exists.']}, status=status.HTTP_400_BAD_REQUEST)

    user = User.objects.create_user(email=email, password=password1)
    token, _ = Token.objects.get_or_create(user=user)
    return Response({'key': token.key}, status=status.HTTP_201_CREATED)


@api_view(['GET'])
@permission_classes([permissions.IsAuthenticated])
def user_stats(request):
    """
    Return basic user statistics.
    """
    user = request.user
    stats = {
        'user_id': user.pk,

        'email': user.email,
        'date_joined': user.date_joined,
        'last_login': user.last_login,
        'is_staff': user.is_staff,
        'is_active': user.is_active,
    }
    return Response(stats)
