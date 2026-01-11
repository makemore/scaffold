from authtools.models import AbstractEmailUser
from django.db import models


# Create your models here.


class User(AbstractEmailUser):
    """
    Custom User model using email as the identifier.
    Includes optional name fields used by tests and serializers.
    """
    first_name = models.CharField('first name', max_length=150, blank=True)
    last_name = models.CharField('last name', max_length=150, blank=True)
    full_name = models.CharField('full name', max_length=255, blank=True)
    preferred_name = models.CharField('preferred name', max_length=255, blank=True)

    def get_full_name(self):
        """
        Return the user's full name.
        """
        return self.full_name.strip() if self.full_name else ''

    def get_short_name(self):
        """
        Return the user's preferred name or first part of full name.
        """
        if self.preferred_name:
            return self.preferred_name.strip()
        elif self.full_name:
            return self.full_name.split()[0] if self.full_name.split() else ''
        return ''

    def __str__(self):
        """
        String representation of the user.
        """
        return self.email

    class Meta:
        verbose_name = 'User'
        verbose_name_plural = 'Users'
