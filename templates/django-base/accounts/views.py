# Views have been moved to accounts/api/views.py
# This file is kept for backward compatibility

from .api.views import UserProfileView, ChangePasswordView, user_stats

__all__ = ['UserProfileView', 'ChangePasswordView', 'user_stats']
