"""
URL configuration for {{ project_slug }} project.

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/5.1/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""
from django.contrib import admin
from django.urls import path, include

urlpatterns = [
    path('admin/', admin.site.urls),

    # All accounts/authentication endpoints under /api/accounts/
    path('api/accounts/', include([
        # dj-rest-auth endpoints
        path('auth/', include('dj_rest_auth.urls')),
        # Override registration to a simple email/password endpoint
        path('auth/registration/', include(([
            path('', __import__('accounts.api.views', fromlist=['']).simple_register, name='rest_register'),
        ], 'accounts'), namespace='auth_registration')),

        path('auth/registration/', include('dj_rest_auth.registration.urls')),

        # Custom accounts endpoints
        path('', include('accounts.api.urls', namespace='accounts')),
    ])),
]
