**To synchronize an MFA device**

The following ``resync-mfa-device`` example synchronizes the MFA device that is associated with the IAM user ``Bob`` and whose ARN is ``arn:aws:iam::123456789012:mfa/BobsMFADevice`` with an authenticator program that provided the two authentication codes. ::

    aws iam resync-mfa-device \
        --user-name Bob \
        --serial-number arn:aws:iam::210987654321:mfa/BobsMFADevice \
        --authentication-code1 123456 \
        --authentication-code2 987654

This command produces no output.

For more information, see `Using Multi-Factor Authentication (MFA) Devices in AWS <http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa.html>`__ in the *AWS Identity and Access Management User Guide*.