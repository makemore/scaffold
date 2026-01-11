# Serializers have been moved to accounts/api/serializers.py
# This file is kept for backward compatibility

from .api.serializers import (
    CustomRegisterSerializer,
    UserDetailsSerializer,
    UserProfileSerializer,
    ChangePasswordSerializer,
)

__all__ = [
    'CustomRegisterSerializer',
    'UserDetailsSerializer',
    'UserProfileSerializer',
    'ChangePasswordSerializer',
]
