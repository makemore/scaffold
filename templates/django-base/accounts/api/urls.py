from django.urls import path
from . import views

app_name = 'accounts'

urlpatterns = [
    path('auth/registration/', views.simple_register, name='rest_register'),
    path('profile/', views.UserProfileView.as_view(), name='user-profile'),
    path('change-password/', views.ChangePasswordView.as_view(), name='change-password'),
    path('stats/', views.user_stats, name='user-stats'),
]
