import os
import unittest
from django.test import TestCase
from django.conf import settings


def s3_is_configured():
    """Check if S3 storage is configured (AWS credentials are set)."""
    return bool(
        getattr(settings, 'AWS_ACCESS_KEY_ID', None) and
        getattr(settings, 'AWS_SECRET_ACCESS_KEY', None) and
        getattr(settings, 'AWS_STORAGE_BUCKET_NAME', None)
    )


@unittest.skipUnless(s3_is_configured(), "S3 not configured (no AWS credentials)")
class S3IntegrationTests(TestCase):
    def test_storages_backend_is_s3(self):
        """Ensure the default storage backend is configured to S3."""
        from django.core.files.storage import default_storage
        try:
            # Newer django-storages exposes S3Storage here
            from storages.backends.s3 import S3Storage  # type: ignore
        except Exception:  # pragma: no cover - fallback for older versions
            # Older versions expose S3Boto3Storage
            from storages.backends.s3boto3 import S3Boto3Storage as S3Storage  # type: ignore

        self.assertTrue(
            isinstance(default_storage, S3Storage),
            msg="Default storage should be an instance of S3Storage/S3Boto3Storage",
        )

    def test_s3_settings_present(self):
        """Verify critical S3 settings are present and sane."""
        # STORAGES mapping should point default to S3 backend
        self.assertIn("default", settings.STORAGES)
        self.assertEqual(
            settings.STORAGES["default"]["BACKEND"],
            "storages.backends.s3.S3Storage",
        )

        # Required AWS settings
        self.assertTrue(settings.AWS_ACCESS_KEY_ID)
        self.assertTrue(settings.AWS_SECRET_ACCESS_KEY)
        self.assertTrue(settings.AWS_STORAGE_BUCKET_NAME)
        self.assertTrue(settings.AWS_S3_ENDPOINT_URL)

        # Security and URL behavior
        self.assertTrue(settings.AWS_QUERYSTRING_AUTH)
        self.assertEqual(settings.AWS_S3_SIGNATURE_VERSION, "s3v4")
        self.assertIn(settings.AWS_S3_ADDRESSING_STYLE, ["auto", "virtual", "path"])


class CloudTasksIntegrationTests(TestCase):
    def test_cloud_tasks_settings_present(self):
        """Basic sanity check for configured django-cloud-tasks settings."""
        self.assertTrue(hasattr(settings, "DJANGO_CLOUD_TASKS"))
        config = settings.DJANGO_CLOUD_TASKS

        # Check required keys exist
        self.assertIn('eager', config)
        self.assertIn('queues', config)
        self.assertIn('default_queue', config)

        # Check all 4 queues are configured
        queues = config['queues']
        self.assertIn('instant', queues)
        self.assertIn('high', queues)
        self.assertIn('medium', queues)
        self.assertIn('low', queues)

        # Default queue should be one of the configured queues
        self.assertIn(config['default_queue'], queues)

    def test_eager_mode_in_development(self):
        """In development, eager mode should be enabled by default."""
        # This test assumes we're running in dev mode where eager=True
        config = settings.DJANGO_CLOUD_TASKS
        # In test environment, eager should be True (tasks run in-process)
        self.assertTrue(config.get('eager', False))

# Live S3 read/write integration tests (skipped unless S3_LIVE_TESTS=1)
import uuid
from django.core.files.base import ContentFile
from django.core.files.storage import default_storage


@unittest.skipUnless(os.getenv("S3_LIVE_TESTS") == "1", "Set S3_LIVE_TESTS=1 to run live S3 tests")
class S3ReadWriteTests(TestCase):
    def test_upload_and_download_small_text_file(self):
        """Upload a small text file to S3 and download it back via default_storage."""
        key = f"test/integration/{uuid.uuid4().hex}.txt"
        content = b"hello from integration test"
        saved_key = None
        try:
            saved_key = default_storage.save(key, ContentFile(content))
            self.assertTrue(default_storage.exists(saved_key))

            with default_storage.open(saved_key, mode="rb") as fh:
                data = fh.read()
            self.assertEqual(data, content)
        finally:
            if saved_key and default_storage.exists(saved_key):
                default_storage.delete(saved_key)

    def test_upload_and_download_binary_file(self):
        """Upload/download a small binary blob to ensure binary IO works as expected."""
        key = f"test/integration/{uuid.uuid4().hex}.bin"
        # Arbitrary binary payload
        content = bytes([0x00, 0xFF, 0x10, 0x20, 0x7F, 0x80, 0xAB, 0xCD])
        saved_key = None
        try:
            saved_key = default_storage.save(key, ContentFile(content))
            self.assertTrue(default_storage.exists(saved_key))

            with default_storage.open(saved_key, mode="rb") as fh:
                data = fh.read()
            self.assertEqual(data, content)
        finally:
            if saved_key and default_storage.exists(saved_key):
                default_storage.delete(saved_key)

