# URLs have been moved to accounts/api/urls.py
# This file is kept for backward compatibility

from django.urls import path, include

urlpatterns = [
    path('', include('accounts.api.urls')),
]
