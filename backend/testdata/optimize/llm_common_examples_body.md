Example 1: When the user clicks pay, call PaymentService.Charge with card token and amount.
Example 2: When charge fails, return 402 and log the provider error code.
Example 3: When charge succeeds, emit PaymentCompleted event to the outbox table.
Example 4: Repeat the same three-step validation for every gateway adapter in adapters/.
Example 5: Document each adapter with integration test fixtures under testdata/payments/.
Example 6: Keep webhook handlers idempotent using dedupe keys stored in Redis.
Example 7: Validate currency codes against the ISO list before submitting charges.
Example 8: Emit audit events for every manual refund initiated by support staff.
