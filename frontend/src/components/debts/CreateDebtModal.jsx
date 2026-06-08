import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useDebtsStore } from '@stores/debtsStore'
import { useContactsStore } from '@stores/contactsStore'
import { useNotificationsStore } from '@stores/notificationsStore'
import { convertToISO } from '@utils/formatters'
import { PaymentFields } from './PaymentFields'
import { CreateContactModal } from '../contacts/CreateContactModal'

export const CreateDebtModal = ({ onDebtCreated, onClose }) => {
  const createDebt = useDebtsStore((state) => state.createDebt)
  const { contacts, fetchContacts } = useContactsStore()
  const { success, error } = useNotificationsStore()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [mouseDownOnOverlay, setMouseDownOnOverlay] = useState(false)
  const [showCreateContactModal, setShowCreateContactModal] = useState(false)
  const [notificationsEnabled, setNotificationsEnabled] = useState(true)

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm({
    defaultValues: {
      contact_id: '',
      debt_type: 'to_pay',
      payment_type: 'onetime',
      total_amount: '',
      description: '',
      due_date: '',
      notes: '',
      number_of_payments: '',
      installment_amount: '',
      installment_plan: 'monthly',
      next_payment_date: '',
      installment_calculation_method: 'by_count', // 'by_count' or 'by_date'
    },
  })

  useEffect(() => {
    fetchContacts()
  }, [])

  const onSubmit = async (data) => {
    setIsSubmitting(true)
    try {
      // Convert due_date to ISO 8601 datetime format if provided
      const debtData = {
        ...data,
        total_amount: String(data.total_amount),
      }

      const convertedDueDate = convertToISO(debtData.due_date)
      if (convertedDueDate) {
        debtData.due_date = convertedDueDate
      } else {
        delete debtData.due_date
      }

      // Handle installment-specific fields
      if (data.payment_type === 'installment') {
        debtData.number_of_payments = parseInt(data.number_of_payments)
        debtData.installment_plan = data.installment_plan

        const convertedNextPaymentDate = convertToISO(data.next_payment_date)
        if (convertedNextPaymentDate) {
          debtData.next_payment_date = convertedNextPaymentDate
        } else {
          delete debtData.next_payment_date
        }
      } else {
        // For one-time payments, set installment_plan to 'onetime' and number_of_payments to 1
        debtData.installment_plan = 'onetime'
        debtData.number_of_payments = 1
        delete debtData.installment_amount
        delete debtData.next_payment_date
      }

      debtData.notifications_enabled = notificationsEnabled
      await createDebt(debtData)
      success('Debt created successfully')
      onDebtCreated()
    } catch (err) {
      error(err.message || 'Failed to create debt')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleOverlayMouseDown = (e) => {
    // Track if mousedown happened on the overlay
    if (e.target === e.currentTarget && !isSubmitting) {
      setMouseDownOnOverlay(true)
    }
  }

  const handleOverlayClick = (e) => {
    // Only close if both mousedown and click happened on the overlay
    if (e.target === e.currentTarget && !isSubmitting && mouseDownOnOverlay) {
      onClose()
    }
    // Reset the flag
    setMouseDownOnOverlay(false)
  }

  const handleContactCreated = async () => {
    // Close the create contact modal
    setShowCreateContactModal(false)
    // Fetch updated contacts list
    await fetchContacts()
    // The newly created contact will be available in the contacts list
    // and will be automatically selected if it's the only contact
  }

  return (
    <div
      className="fixed inset-0 z-50 !mt-0 flex items-start justify-center overflow-y-auto bg-black/60 p-4 pb-96"
      onMouseDown={handleOverlayMouseDown}
      onClick={handleOverlayClick}
    >
      <div className="card my-8 w-full max-w-lg overflow-visible">
        <div className="border-b border-border px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-foreground">Create New Debt</h2>
            <button
              onClick={onClose}
              disabled={isSubmitting}
              className="text-muted-foreground transition-colors hover:text-foreground disabled:cursor-not-allowed"
            >
              <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="p-6">
          <div className="space-y-4">
            {/* Debt Type */}
            <div>
              <label htmlFor="debt_type" className="mb-2 block text-sm font-medium text-foreground">
                Debt Type <span className="text-destructive">*</span>
              </label>
              <select
                id="debt_type"
                {...register('debt_type', { required: 'Debt type is required' })}
                className="input"
                disabled={isSubmitting}
              >
                <option value="to_pay">To Pay</option>
                <option value="to_receive">To Receive</option>
              </select>
              {errors.debt_type && (
                <p className="mt-1 text-sm text-destructive">{errors.debt_type.message}</p>
              )}
            </div>

            {/* Contact */}
            <div>
              <label
                htmlFor="contact_id"
                className="mb-2 block text-sm font-medium text-foreground"
              >
                Contact <span className="text-destructive">*</span>
              </label>
              <div className="space-y-2">
                <select
                  id="contact_id"
                  {...register('contact_id', { required: 'Contact is required' })}
                  className="input"
                  disabled={isSubmitting}
                >
                  <option value="">Select a contact</option>
                  {contacts.map((contact) => (
                    <option key={contact.id} value={contact.id}>
                      {contact.name}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  onClick={() => setShowCreateContactModal(true)}
                  disabled={isSubmitting}
                  className="flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                    />
                  </svg>
                  Add New Contact
                </button>
              </div>
              {errors.contact_id && (
                <p className="mt-1 text-sm text-destructive">{errors.contact_id.message}</p>
              )}
            </div>

            {/* Total Amount */}
            <div>
              <label
                htmlFor="total_amount"
                className="mb-2 block text-sm font-medium text-foreground"
              >
                Total Amount <span className="text-destructive">*</span>
              </label>
              <input
                id="total_amount"
                type="number"
                step="0.01"
                min="0"
                {...register('total_amount', {
                  required: 'Total amount is required',
                  min: { value: 0.01, message: 'Amount must be greater than 0' },
                })}
                className="input"
                disabled={isSubmitting}
                placeholder="0.00"
              />
              {errors.total_amount && (
                <p className="mt-1 text-sm text-destructive">{errors.total_amount.message}</p>
              )}
            </div>

            {/* Payment Fields Component */}
            <PaymentFields
              register={register}
              watch={watch}
              setValue={setValue}
              errors={errors}
              isSubmitting={isSubmitting}
            />

            {/* Description */}
            <div>
              <label
                htmlFor="description"
                className="mb-2 block text-sm font-medium text-foreground"
              >
                Description
              </label>
              <input
                id="description"
                type="text"
                {...register('description', {
                  maxLength: {
                    value: 255,
                    message: 'Description must be less than 255 characters',
                  },
                })}
                className="input"
                disabled={isSubmitting}
                placeholder="What is this debt for?"
              />
              {errors.description && (
                <p className="mt-1 text-sm text-destructive">{errors.description.message}</p>
              )}
            </div>

            {/* Notes */}
            <div>
              <label htmlFor="notes" className="mb-2 block text-sm font-medium text-foreground">
                Notes
              </label>
              <textarea
                id="notes"
                rows={3}
                {...register('notes')}
                className="input resize-none"
                disabled={isSubmitting}
                placeholder="Additional notes or details"
              />
              {errors.notes && (
                <p className="mt-1 text-sm text-destructive">{errors.notes.message}</p>
              )}
            </div>

            {/* Notifications toggle */}
            <div className="flex items-center justify-between rounded-lg border border-border p-4">
              <div>
                <p className="text-sm font-medium text-foreground">Enable payment reminders</p>
                <p className="text-xs text-muted-foreground">
                  Automatically schedule reminder notifications based on due dates
                </p>
              </div>
              <button
                type="button"
                onClick={() => setNotificationsEnabled((v) => !v)}
                disabled={isSubmitting}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                  notificationsEnabled ? 'bg-primary' : 'bg-muted'
                }`}
                aria-label={notificationsEnabled ? 'Disable reminders' : 'Enable reminders'}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow ${
                    notificationsEnabled ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>
          </div>

          <div className="mt-6 flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className="btn-secondary flex-1"
            >
              Cancel
            </button>
            <button type="submit" disabled={isSubmitting} className="btn-primary flex-1">
              {isSubmitting ? (
                <span className="flex items-center justify-center">
                  <svg className="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24">
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                      fill="none"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Creating...
                </span>
              ) : (
                'Create Debt'
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Create Contact Modal */}
      {showCreateContactModal && (
        <CreateContactModal
          onContactCreated={handleContactCreated}
          onClose={() => setShowCreateContactModal(false)}
        />
      )}
    </div>
  )
}
