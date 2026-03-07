# TASK 1.1.13: Mailer Integration Test (Testcontainers)

## Status: ✅ done

**Last Updated:** 2026-03-07

---

## 📖 Description

Implement a true integration test for the `SMTPMailer` using `testcontainers-go` and `Mailhog`. This ensures that the generated email format is correctly parsed by a real SMTP server and can be retrieved via API.

---

## 🎯 Acceptance Criteria

- [x] Spin up a `Mailhog` container using `testcontainers-go`.
- [x] Send an OTP using `SMTPMailer.SendOTP`.
- [x] Verify the email arrived in Mailhog via its HTTP API.
- [x] Ensure the container is cleaned up after tests.
- [x] `make task-check` passes.

---

## 🛠️ Technical Notes

- Use `github.com/testcontainers/testcontainers-go`.
- Use `net/http` to query Mailhog's `/api/v2/messages`.

---

## 📝 Change Log

- 2026-03-07: Initial task creation.
- 2026-03-07: Task completed. Integration test passed using testcontainers and Mailhog.
