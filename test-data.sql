-- ============================================
-- DEVELOPMENT SEED DATA
-- Loaded automatically when APP_ENV=development
-- ============================================
-- Demo logins (bcrypt-hashed passwords):
--   alice@dev.local  / password123  (primary demo user, full debt scenarios)
--   bob@dev.local    / password123  (creditor with cross-user debts)
--   carol@dev.local  / devpassword  (minimal SMS-focused settings)
-- ============================================

-- Users, contacts, and custom settings
DO $$
DECLARE
    v_now timestamp := NOW();
    v_password_hash text := '$2a$10$ZUW.p9HtGPO09t5Bwqaqi.20vurD3773.RJ9m1mXAMs8rKfksaZx2';
    v_carol_password_hash text := '$2a$10$.i7WLuynHOvFvQw7GYzIneWEHEpc0SgVYLO62ZdMnCqHthtVSz9Pa';

    v_alice_id uuid := '10000000-0000-0000-0000-000000000001';
    v_bob_id uuid := '10000000-0000-0000-0000-000000000002';
    v_carol_id uuid := '10000000-0000-0000-0000-000000000003';

    v_maria_contact_id uuid := '30000000-0000-0000-0000-000000000001';
    v_bob_contact_id uuid := '30000000-0000-0000-0000-000000000002';
    v_landlord_contact_id uuid := '30000000-0000-0000-0000-000000000003';
    v_alice_contact_id uuid := '30000000-0000-0000-0000-000000000004';
    v_supplier_contact_id uuid := '30000000-0000-0000-0000-000000000005';
BEGIN
    INSERT INTO users (id, email, password_hash, first_name, last_name, phone, created_at, updated_at)
    VALUES
        (v_alice_id, 'alice@dev.local', v_password_hash, 'Alice', 'Demo', '+639171234567', v_now, v_now),
        (v_bob_id, 'bob@dev.local', v_password_hash, 'Bob', 'Creditor', '+639181234567', v_now, v_now),
        (v_carol_id, 'carol@dev.local', v_carol_password_hash, 'Carol', 'Minimal', '+639191234567', v_now, v_now);

    INSERT INTO user_settings (
        id, user_id,
        notification_email, notification_sms, notification_webhook,
        notification_reminder_days, notification_time, overdue_reminder_frequency,
        custom_email_message, custom_sms_message,
        slack_webhook_url, telegram_chat_id, discord_webhook_url,
        event_notifications_enabled, notify_contact_on_payment, notification_recipient,
        default_currency, timezone, created_at, updated_at
    ) VALUES
        (
            '20000000-0000-0000-0000-000000000001', v_alice_id,
            true, true, true,
            ARRAY[14, 7, 3, 1]::integer[], '08:30:00', 'daily',
            'Hi! Friendly reminder from Alice about your upcoming payment for {{debt_description}}.',
            'Payment reminder: {{amount}} due {{due_date}}. Reply PAID once sent.',
            'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
            '123456789', 'https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX',
            true, true, 'both',
            'Php', 'Asia/Manila', v_now, v_now
        ),
        (
            '20000000-0000-0000-0000-000000000002', v_bob_id,
            true, false, false,
            ARRAY[7, 1]::integer[], '09:00:00', 'weekly',
            NULL, NULL,
            NULL, NULL, NULL,
            true, false, 'user',
            'Php', 'UTC', v_now, v_now
        ),
        (
            '20000000-0000-0000-0000-000000000003', v_carol_id,
            false, true, true,
            ARRAY[3, 1]::integer[], '18:00:00', 'daily',
            NULL, 'Carol here - {{amount}} is due on {{due_date}}.',
            NULL, NULL, 'https://discord.com/api/webhooks/111111111111111111/YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY',
            false, true, 'contact',
            'USD', 'America/New_York', v_now, v_now
        );

    INSERT INTO contacts (id, is_user, user_id_ref, created_at, updated_at)
    VALUES
        (v_maria_contact_id, false, NULL, v_now, v_now),
        (v_bob_contact_id, true, v_bob_id, v_now, v_now),
        (v_landlord_contact_id, false, NULL, v_now, v_now),
        (v_alice_contact_id, true, v_alice_id, v_now, v_now),
        (v_supplier_contact_id, false, NULL, v_now, v_now);

    INSERT INTO user_contacts (id, user_id, contact_id, name, email, phone, notes, created_at, updated_at)
    VALUES
        ('40000000-0000-0000-0000-000000000001', v_alice_id, v_maria_contact_id, 'Maria Santos', 'maria.santos@example.com', '+639201112233', 'Roommate and frequent lunch split partner', v_now, v_now),
        ('40000000-0000-0000-0000-000000000002', v_alice_id, v_bob_contact_id, 'Bob Creditor', 'bob@dev.local', '+639181234567', 'Registered app user - linked contact', v_now, v_now),
        ('40000000-0000-0000-0000-000000000003', v_alice_id, v_landlord_contact_id, 'Office Landlord', 'landlord@example.com', '+63281234567', 'Monthly office rent', v_now, v_now),
        ('40000000-0000-0000-0000-000000000004', v_bob_id, v_alice_contact_id, 'Alice Demo', 'alice@dev.local', '+639171234567', 'Owes Bob for shared equipment purchase', v_now, v_now),
        ('40000000-0000-0000-0000-000000000005', v_bob_id, v_supplier_contact_id, 'Supplier Co.', 'billing@supplier.example.com', '+63289998877', 'Inventory vendor', v_now, v_now);

    RAISE NOTICE 'Seeded 3 users, 3 custom settings profiles, and 5 contacts';
END $$;

-- Debt scenarios for Alice (Maria Santos) and Bob (Alice Demo)
DO $$
DECLARE
    v_alice_id uuid := '10000000-0000-0000-0000-000000000001';
    v_bob_id uuid := '10000000-0000-0000-0000-000000000002';
    v_maria_contact_id uuid := '30000000-0000-0000-0000-000000000001';
    v_alice_contact_id uuid := '30000000-0000-0000-0000-000000000004';

    v_user_id uuid;
    v_contact_id uuid;
    v_debt_id uuid;
    v_created_date timestamp;
    v_due_date timestamp;
    v_next_payment_date timestamp;
BEGIN
    v_user_id := v_alice_id;
    v_contact_id := v_maria_contact_id;

    -- Test Case 1: WEEKLY - Overdue with mixed payments
    v_debt_id := '11111111-1111-1111-1111-111111111111'::uuid;
    v_created_date := NOW() - INTERVAL '70 days';
    v_due_date := v_created_date + INTERVAL '70 days';
    v_next_payment_date := v_created_date + INTERVAL '28 days';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        500.00, 50.00, 150.00, 350.00, 'Php', 'overdue',
        v_due_date, v_next_payment_date,
        'weekly', 10,
        'Weekly lunch split with Maria - overdue',
        'Includes failed week 4 transfer', v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        payment_data.payment_method,
        payment_data.description,
        payment_data.status,
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (50.00, v_created_date + INTERVAL '7 days', 'bank_transfer', 'Week 1 payment', 'completed'),
        (50.00, v_created_date + INTERVAL '14 days', 'cash', 'Week 2 payment', 'completed'),
        (50.00, v_created_date + INTERVAL '21 days', 'digital_wallet', 'Week 3 payment', 'completed'),
        (50.00, v_created_date + INTERVAL '28 days', 'bank_transfer', 'Week 4 payment - FAILED', 'failed')
    ) AS payment_data(amount, payment_date, payment_method, description, status);

    -- Test Case 2: MONTHLY - Active with some payments
    v_debt_id := '22222222-2222-2222-2222-222222222222'::uuid;
    v_created_date := NOW() - INTERVAL '3 months' - INTERVAL '28 days';
    v_due_date := v_created_date + INTERVAL '12 months';
    v_next_payment_date := v_created_date + INTERVAL '4 months';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_receive',
        1200.00, 100.00, 300.00, 900.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'monthly', 12,
        'Concert tickets Maria is paying back',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        payment_data.payment_method,
        payment_data.description,
        'completed',
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (100.00, v_created_date + INTERVAL '1 month', 'bank_transfer', 'Month 1 payment'),
        (100.00, v_created_date + INTERVAL '2 months', 'bank_transfer', 'Month 2 payment'),
        (100.00, v_created_date + INTERVAL '3 months', 'cash', 'Month 3 payment')
    ) AS payment_data(amount, payment_date, payment_method, description);

    -- Test Case 3: BIWEEKLY - Overdue with pending payment
    v_debt_id := '33333333-3333-3333-3333-333333333333'::uuid;
    v_created_date := NOW() - INTERVAL '112 days';
    v_due_date := v_created_date + INTERVAL '112 days';
    v_next_payment_date := v_created_date + INTERVAL '42 days';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        800.00, 100.00, 200.00, 600.00, 'Php', 'overdue',
        v_due_date, v_next_payment_date,
        'biweekly', 8,
        'Shared utility bills - overdue with pending installment',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        payment_data.payment_method,
        payment_data.description,
        payment_data.status,
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (100.00, v_created_date + INTERVAL '14 days', 'bank_transfer', 'Biweekly 1', 'completed'),
        (100.00, v_created_date + INTERVAL '28 days', 'cash', 'Biweekly 2', 'completed'),
        (100.00, v_created_date + INTERVAL '42 days', 'digital_wallet', 'Biweekly 3 - PENDING', 'pending')
    ) AS payment_data(amount, payment_date, payment_method, description, status);

    -- Test Case 4: QUARTERLY - Active, early payments
    v_debt_id := '44444444-4444-4444-4444-444444444444'::uuid;
    v_created_date := NOW() - INTERVAL '3 months';
    v_due_date := v_created_date + INTERVAL '12 months';
    v_next_payment_date := v_created_date + INTERVAL '6 months';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_receive',
        2000.00, 500.00, 500.00, 1500.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'quarterly', 4,
        'Group trip deposit Maria owes',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        v_debt_id,
        500.00,
        'Php',
        v_created_date + INTERVAL '3 months',
        'bank_transfer',
        'Quarter 1 payment',
        'completed',
        v_created_date + INTERVAL '3 months',
        v_created_date + INTERVAL '3 months'
    );

    -- Test Case 5: ONETIME - Overdue, no payments
    v_debt_id := '55555555-5555-5555-5555-555555555555'::uuid;
    v_created_date := NOW() - INTERVAL '60 days';
    v_due_date := NOW() - INTERVAL '30 days';
    v_next_payment_date := v_due_date;

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        5000.00, 5000.00, 0.00, 5000.00, 'Php', 'overdue',
        v_due_date, v_next_payment_date,
        'onetime', 1,
        'Emergency laptop repair loan',
        'No payments recorded yet', v_created_date, NOW()
    );

    -- Test Case 6: YEARLY - Active, long term
    v_debt_id := '66666666-6666-6666-6666-666666666666'::uuid;
    v_created_date := NOW() - INTERVAL '2 years' - INTERVAL '1 month';
    v_due_date := v_created_date + INTERVAL '5 years';
    v_next_payment_date := v_created_date + INTERVAL '3 years';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        10000.00, 2000.00, 4000.00, 6000.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'yearly', 5,
        'Annual membership fee split',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        'bank_transfer',
        payment_data.description,
        'completed',
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (2000.00, v_created_date + INTERVAL '1 year', 'Year 1 payment'),
        (2000.00, v_created_date + INTERVAL '2 years', 'Year 2 payment')
    ) AS payment_data(amount, payment_date, description);

    -- Test Case 7: MONTHLY - Partially paid, overdue
    v_debt_id := '77777777-7777-7777-7777-777777777777'::uuid;
    v_created_date := NOW() - INTERVAL '5 months';
    v_due_date := NOW();
    v_next_payment_date := v_created_date + INTERVAL '3 months';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_receive',
        1000.00, 200.00, 550.00, 450.00, 'Php', 'overdue',
        v_due_date, v_next_payment_date,
        'monthly', 5,
        'Furniture installment from Maria',
        'Includes one partial payment', v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        payment_data.payment_method,
        payment_data.description,
        'completed',
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (200.00, v_created_date + INTERVAL '1 month', 'bank_transfer', 'Month 1 payment'),
        (200.00, v_created_date + INTERVAL '2 months', 'cash', 'Month 2 payment'),
        (150.00, v_created_date + INTERVAL '3 months', 'digital_wallet', 'Month 3 partial payment')
    ) AS payment_data(amount, payment_date, payment_method, description);

    -- Test Case 8: WEEKLY - Near completion
    v_debt_id := '88888888-8888-8888-8888-888888888888'::uuid;
    v_created_date := NOW() - INTERVAL '56 days';
    v_due_date := NOW();
    v_next_payment_date := v_created_date + INTERVAL '56 days';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        400.00, 50.00, 350.00, 50.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'weekly', 8,
        'Coffee run tab - almost complete',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        50.00,
        'Php',
        v_created_date + (INTERVAL '1 day' * n * 7),
        CASE (n % 3)
            WHEN 0 THEN 'bank_transfer'
            WHEN 1 THEN 'cash'
            ELSE 'digital_wallet'
        END,
        'completed',
        v_created_date + (INTERVAL '1 day' * n * 7),
        v_created_date + (INTERVAL '1 day' * n * 7)
    FROM generate_series(1, 7) AS n;

    -- Test Case 9: SETTLED debt (fully paid)
    v_debt_id := '99999999-9999-9999-9999-999999999999'::uuid;
    v_created_date := NOW() - INTERVAL '60 days';
    v_due_date := NOW() - INTERVAL '15 days';
    v_next_payment_date := v_due_date;

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        300.00, 100.00, 300.00, 0.00, 'Php', 'settled',
        v_due_date, v_next_payment_date,
        'monthly', 3,
        'Phone case and accessories',
        'Fully settled', v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        v_debt_id,
        payment_data.amount,
        'Php',
        payment_data.payment_date,
        payment_data.payment_method,
        payment_data.description,
        'completed',
        payment_data.payment_date,
        payment_data.payment_date
    FROM (VALUES
        (100.00, v_created_date + INTERVAL '1 month', 'bank_transfer', 'Final payment 1/3'),
        (100.00, v_created_date + INTERVAL '2 months', 'cash', 'Final payment 2/3'),
        (100.00, v_created_date + INTERVAL '45 days', 'digital_wallet', 'Final payment 3/3')
    ) AS payment_data(amount, payment_date, payment_method, description);

    -- Test Case 10: ARCHIVED debt
    v_debt_id := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid;
    v_created_date := NOW() - INTERVAL '1 year';
    v_due_date := NOW() - INTERVAL '6 months';
    v_next_payment_date := v_due_date;

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_receive',
        1000.00, 1000.00, 0.00, 1000.00, 'Php', 'settled',
        v_due_date, v_next_payment_date,
        'onetime', 1,
        'Old project invoice - archived',
        'User archived this debt - no longer pursuing payment', v_created_date, NOW()
    );

    -- Bob's debts with Alice (cross-user scenario)
    v_user_id := v_bob_id;
    v_contact_id := v_alice_contact_id;

    v_debt_id := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1'::uuid;
    v_created_date := NOW() - INTERVAL '45 days';
    v_due_date := v_created_date + INTERVAL '90 days';
    v_next_payment_date := v_created_date + INTERVAL '30 days';

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_receive',
        1500.00, 500.00, 500.00, 1000.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'monthly', 3,
        'Shared equipment purchase - Alice owes Bob',
        NULL, v_created_date, NOW()
    );

    INSERT INTO debt_items (id, debt_list_id, amount, currency, payment_date, payment_method, description, status, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        v_debt_id,
        500.00,
        'Php',
        v_created_date + INTERVAL '15 days',
        'bank_transfer',
        'First installment from Alice',
        'completed',
        v_created_date + INTERVAL '15 days',
        v_created_date + INTERVAL '15 days'
    );

    v_debt_id := 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2'::uuid;
    v_created_date := NOW() - INTERVAL '20 days';
    v_due_date := NOW() + INTERVAL '10 days';
    v_next_payment_date := v_due_date;

    INSERT INTO debt_lists (
        id, user_id, contact_id, debt_type, total_amount, installment_amount,
        total_payments_made, total_remaining_debt, currency, status,
        due_date, next_payment_date, installment_plan, number_of_payments,
        description, notes, created_at, updated_at
    ) VALUES (
        v_debt_id, v_user_id, v_contact_id, 'to_pay',
        750.00, 750.00, 0.00, 750.00, 'Php', 'active',
        v_due_date, v_next_payment_date,
        'onetime', 1,
        'Alice fronted conference tickets for Bob',
        'Due soon', v_created_date, NOW()
    );

    RAISE NOTICE 'Seeded 12 debt lists with payment history for development';
END $$;
