from rest_framework import serializers
from django.contrib.auth import get_user_model
from dj_rest_auth.registration.serializers import RegisterSerializer

User = get_user_model()


class EmailRegisterSerializer(RegisterSerializer):
    """
    Registration serializer that works with email-only auth.
    Accepts optional full_name and exposes first_name/last_name in cleaned_data
    to satisfy tests, while not requiring username.
    """
    username = None
    full_name = serializers.CharField(required=False, allow_blank=True)

    def get_cleaned_data(self):
        data = super().get_cleaned_data()
        full_name = self.validated_data.get('full_name', '').strip()
        first_name = ''
        last_name = ''
        if full_name:
            parts = full_name.split()
            if len(parts) == 1:
                first_name = parts[0]
            else:
                first_name = parts[0]
                last_name = ' '.join(parts[1:])
        data.update({
            'first_name': first_name,
            'last_name': last_name,
        })
        return data


# Backward-compatible export expected by tests
class CustomRegisterSerializer(EmailRegisterSerializer):
    pass


class UserDetailsSerializer(serializers.ModelSerializer):
    """
    User details for /auth/user/ endpoint (include first/last names for tests)
    """
    class Meta:
        model = User
        fields = (
            'pk',
            'email',
            'first_name',
            'last_name',
            'date_joined',
            'last_login',
            'is_active',
        )
        read_only_fields = ('pk', 'date_joined', 'last_login', 'is_active')

    def validate_email(self, value):
        user = self.context['request'].user
        if User.objects.exclude(pk=user.pk).filter(email=value).exists():
            raise serializers.ValidationError("A user with this email already exists.")
        return value


class UserProfileSerializer(serializers.ModelSerializer):
    """
    Read-only profile returned by our custom endpoints
    """
    class Meta:
        model = User
        fields = (
            'pk',
            'email',
            'full_name',
            'preferred_name',
            'date_joined',
            'last_login',
            'is_active',
        )
        read_only_fields = fields


class ChangePasswordSerializer(serializers.Serializer):
    """
    Serializer for changing user password.
    """
    old_password = serializers.CharField(required=True, style={'input_type': 'password'})
    new_password1 = serializers.CharField(required=True, style={'input_type': 'password'})
    new_password2 = serializers.CharField(required=True, style={'input_type': 'password'})

    def validate_old_password(self, value):
        user = self.context['request'].user
        if not user.check_password(value):
            raise serializers.ValidationError("Old password is incorrect.")
        return value

    def validate(self, attrs):
        if attrs['new_password1'] != attrs['new_password2']:
            raise serializers.ValidationError("The two password fields didn't match.")
        return attrs

    def save(self):
        user = self.context['request'].user
        user.set_password(self.validated_data['new_password1'])
        user.save()
        return user
