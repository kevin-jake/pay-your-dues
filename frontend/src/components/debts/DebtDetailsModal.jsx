import { useState, useEffect } from 'react'
import { usePaymentsStore } from '@stores/paymentsStore'
import { useDebtsStore } from '@stores/debtsStore'
import { ImageViewerModal } from '@components/common/ImageViewerModal'
import { PaymentHistory } from './PaymentHistory'
import { InstallmentScheduleModal } from './InstallmentScheduleModal'
import { DebtNotificationsPanel } from './DebtNotificationsPanel'
import {
  formatCurrency,
  formatDate,
  getDebtStatus,
  getDaysUntilDue,
  getDueDateColor,
  convertToISO,
} from '@utils/formatters'

export const DebtDetailsModal = ({ debt, onClose, onEdit, onDelete }) => {
  const {
    fetchPayments,
    createPayment,
    uploadReceipt,
    isLoading: paymentsLoading,
  } = usePaymentsStore()
  const fetchPaymentSchedule = useDebtsStore((state) => state.fetchPaymentSchedule)
  const [debtPayments, setDebtPayments] = useState([])
  const [showAddPayment, setShowAddPayment] = useState(false)
  const [newPayment, setNewPayment] = useState({
    payment_date: new Date().toISOString().split('T')[0],
    amount: '',
    description: '',
  })
  const [receiptFile, setReceiptFile] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [viewingReceipt, setViewingReceipt] = useState(null)
  const [showSchedule, setShowSchedule] = useState(false)
  const [activeTab, setActiveTab] = useState('details')
  const [nextPaymentInfo, setNextPaymentInfo] = useState(null)
  const [loadingSchedule, setLoadingSchedule] = useState(false)

  useEffect(() => {
    loadPayments(debt.id)
    loadNextPaymentInfo(debt.id)
  }, [debt.id])

  const loadPayments = async (debtId) => {
    console.log('debtId', debtId)
    try {
      const paymentsData = await fetchPayments(debtId)
      setDebtPayments(paymentsData || [])
    } catch (error) {
      console.error('Failed to load payments:', error)
      setDebtPayments([])
    }
  }

  const loadNextPaymentInfo = async (debtId) => {
    // Only load if it's an installment debt
    if (!debt.installment_plan || debt.installment_plan === 'onetime') {
      return
    }

    try {
      setLoadingSchedule(true)
      const schedule = await fetchPaymentSchedule(debtId)

      // Find the next payment that is not paid
      const nextPayment = schedule.find((item) => item.status !== 'paid')

      if (nextPayment) {
        // Check if the payment is overdue based on due date
        const dueDate = new Date(nextPayment.due_date)
        const today = new Date()
        today.setHours(0, 0, 0, 0)
        dueDate.setHours(0, 0, 0, 0)

        // Update status if overdue
        if (nextPayment.status === 'pending' && dueDate < today) {
          nextPayment.status = 'overdue'
        }
      }

      setNextPaymentInfo(nextPayment || null)
    } catch (error) {
      console.error('Failed to load next payment info:', error)
      setNextPaymentInfo(null)
    } finally {
      setLoadingSchedule(false)
    }
  }

  const handleAddPayment = async (e) => {
    e.preventDefault()
    if (!newPayment.amount || parseFloat(newPayment.amount) <= 0) {
      alert('Please enter a valid amount')
      return
    }

    try {
      setIsSubmitting(true)
      // Convert date string to ISO 8601 datetime format
      const paymentDateTime = convertToISO(newPayment.payment_date)
      if (!paymentDateTime) {
        alert('Invalid payment date')
        return
      }

      const payment = await createPayment(debt.id, {
        payment_date: paymentDateTime,
        amount: String(newPayment.amount),
        description: newPayment.description,
      })

      // Upload receipt if file is selected
      if (receiptFile && payment.id) {
        await uploadReceipt(payment.id, receiptFile)
      }

      // Reload payments and next payment info
      await loadPayments(payment.debt_list_id)
      await loadNextPaymentInfo(payment.debt_list_id)

      // Reset form
      setNewPayment({
        payment_date: new Date().toISOString().split('T')[0],
        amount: '',
        description: '',
      })
      setReceiptFile(null)
      setShowAddPayment(false)
    } catch (error) {
      console.error('Failed to add payment:', error)
      alert('Failed to add payment. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleFileChange = (e) => {
    const file = e.target.files?.[0]
    if (file) {
      // Validate file type
      const validTypes = ['image/jpeg', 'image/png', 'image/jpg', 'image/webp']
      if (!validTypes.includes(file.type)) {
        alert('Please select a valid image file (JPEG, PNG, or WebP)')
        return
      }
      // Validate file size (max 5MB)
      if (file.size > 5 * 1024 * 1024) {
        alert('File size must be less than 5MB')
        return
      }
      setReceiptFile(file)
    }
  }

  const handleOverlayClick = (e) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  // Calculate total paid and remaining balance
  const totalPaid = debtPayments.reduce((sum, payment) => sum + parseFloat(payment.amount || 0), 0)
  const remainingBalance = parseFloat(debt.total_amount || 0) - totalPaid

  return (
    <div
      className="fixed inset-0 z-50 !mt-0 flex items-start justify-center overflow-y-auto bg-black/60 p-4 pb-96"
      onClick={handleOverlayClick}
    >
      <div className="card my-8 w-full max-w-2xl overflow-visible">
        <div className="border-b border-border px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-foreground">Debt Details</h2>
            <button
              onClick={onClose}
              className="text-muted-foreground transition-colors hover:text-foreground"
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

        <div className="border-b border-border px-6">
          <div className="flex space-x-1">
            <button
              onClick={() => setActiveTab('details')}
              className={`border-b-2 px-4 py-3 text-sm font-medium transition-colors ${
                activeTab === 'details'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              Details
            </button>
            <button
              onClick={() => setActiveTab('notifications')}
              className={`border-b-2 px-4 py-3 text-sm font-medium transition-colors ${
                activeTab === 'notifications'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              Notifications
            </button>
          </div>
        </div>

        <div className="p-6">
          {activeTab === 'notifications' ? (
            <DebtNotificationsPanel debt={debt} />
          ) : (
          <div className="space-y-6">
            {/* Debt Type Badge */}
            <div className="flex items-center justify-between">
              <span
                className={`inline-flex rounded-full px-3 py-1 text-sm font-medium ${
                  debt.debt_type === 'to_pay'
                    ? 'bg-destructive/10 text-destructive'
                    : 'bg-success/10 text-success'
                }`}
              >
                {debt.debt_type === 'to_pay' ? 'To Pay' : 'To Receive'}
              </span>
              <div className="text-sm text-muted-foreground">
                {(() => {
                  const status = getDebtStatus(debt.due_date)
                  return (
                    <span
                      className={`inline-flex rounded-full px-2 py-1 ${status.color} ${status.bgColor}`}
                    >
                      {status.label}
                    </span>
                  )
                })()}
              </div>
            </div>

            {/* Amount Section */}
            <div className="rounded-lg border border-border bg-muted/50 p-6">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                <div className="text-center">
                  <div className="mb-2 text-sm text-muted-foreground">Total Amount</div>
                  <div
                    className={`text-2xl font-bold ${
                      debt.debt_type === 'to_pay' ? 'text-destructive' : 'text-success'
                    }`}
                  >
                    {formatCurrency(parseFloat(debt.total_amount || 0))}
                  </div>
                </div>
                <div className="text-center">
                  <div className="mb-2 text-sm text-muted-foreground">Total Paid</div>
                  <div className="text-2xl font-bold text-primary">{formatCurrency(totalPaid)}</div>
                </div>
                <div className="text-center">
                  <div className="mb-2 text-sm text-muted-foreground">Remaining</div>
                  <div
                    className={`text-2xl font-bold ${
                      remainingBalance > 0
                        ? debt.debt_type === 'to_pay'
                          ? 'text-destructive'
                          : 'text-success'
                        : 'text-muted-foreground'
                    }`}
                  >
                    {formatCurrency(remainingBalance)}
                  </div>
                </div>
              </div>
            </div>

            {/* Contact Information */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div className="rounded-lg border border-border p-4">
                <div className="mb-2 text-sm font-medium text-muted-foreground">Contact</div>
                <div className="flex items-center space-x-3">
                  <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                    <span className="text-lg font-medium text-primary">
                      {debt.contact?.name
                        ?.split(' ')
                        .map((n) => n[0])
                        .join('')
                        .toUpperCase() || 'U'}
                    </span>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      {debt.contact?.name || 'Unknown Contact'}
                    </div>
                    {debt.contact?.email && (
                      <div className="text-sm text-muted-foreground">{debt.contact.email}</div>
                    )}
                    {debt.contact?.phone && (
                      <div className="text-sm text-muted-foreground">{debt.contact.phone}</div>
                    )}
                  </div>
                </div>
              </div>
              {/* Due Date */}
              {debt.due_date && (
                <div className="rounded-lg border border-border p-4">
                  <div className="mb-2 text-sm font-medium text-muted-foreground">Due Date</div>
                  <div className="flex flex-col justify-between">
                    <div className="text-foreground">{formatDate(debt.due_date)}</div>
                    {getDaysUntilDue(debt.due_date) !== null && debt.status !== 'settled' && (
                      <span
                        className={`text-sm font-medium ${getDueDateColor(getDaysUntilDue(debt.due_date))}`}
                      >
                        {getDaysUntilDue(debt.due_date) <= 3
                          ? getDaysUntilDue(debt.due_date) === 0
                            ? 'Due today'
                            : debt.status === 'overdue' || getDaysUntilDue(debt.due_date) < 0
                              ? `Overdue by ${Math.abs(getDaysUntilDue(debt.due_date))} day${Math.abs(getDaysUntilDue(debt.due_date)) === 1 ? '' : 's'}`
                              : `⚠️ Due in ${getDaysUntilDue(debt.due_date)} day${getDaysUntilDue(debt.due_date) === 1 ? '' : 's'}`
                          : ''}
                      </span>
                    )}
                    {debt.status === 'settled' && (
                      <span className="text-sm font-medium text-muted-foreground">Settled</span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Payment Frequency */}
            {debt.installment_plan && (
              <div className="rounded-lg border border-border p-4">
                <div className="mb-2 text-sm font-medium text-muted-foreground">Payment Plan</div>
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-lg font-semibold capitalize text-foreground">
                      {debt.installment_plan === 'onetime'
                        ? 'One-time Payment'
                        : debt.installment_plan === 'biweekly'
                          ? 'Bi-weekly Installments'
                          : `${debt.installment_plan} Installments`}
                    </div>
                    {debt.installment_plan !== 'onetime' && debt.number_of_payments && (
                      <div className="mt-1 text-sm text-muted-foreground">
                        {debt.number_of_payments} payment{debt.number_of_payments !== 1 ? 's' : ''}{' '}
                        total
                      </div>
                    )}
                  </div>
                  <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                    <svg
                      className="h-6 w-6 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                      />
                    </svg>
                  </div>
                </div>
              </div>
            )}

            {/* Next Payment Date */}
            {debt.installment_plan &&
              debt.installment_plan !== 'onetime' &&
              debt.number_of_payments && (
                <div className="rounded-lg border border-border p-4">
                  <div className="mb-2 text-sm font-medium text-muted-foreground">Next Payment</div>
                  {loadingSchedule ? (
                    <div className="py-4 text-center">
                      <div className="border-3 inline-block h-6 w-6 animate-spin rounded-full border-solid border-primary border-r-transparent"></div>
                    </div>
                  ) : nextPaymentInfo ? (
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="text-sm text-muted-foreground">Due Date</div>
                          <div className="text-foreground">
                            {formatDate(nextPaymentInfo.due_date)}
                          </div>
                        </div>
                        {getDaysUntilDue(nextPaymentInfo.due_date) !== null && (
                          <span
                            className={`text-sm font-medium ${getDueDateColor(getDaysUntilDue(nextPaymentInfo.due_date))}`}
                          >
                            {nextPaymentInfo.status === 'overdue' ||
                            nextPaymentInfo.status === 'missed'
                              ? `Overdue by ${Math.abs(getDaysUntilDue(nextPaymentInfo.due_date))} days`
                              : getDaysUntilDue(nextPaymentInfo.due_date) === 0
                                ? 'Due today'
                                : getDaysUntilDue(nextPaymentInfo.due_date) <= 3
                                  ? `⚠️ Due in ${getDaysUntilDue(nextPaymentInfo.due_date)} day${getDaysUntilDue(nextPaymentInfo.due_date) === 1 ? '' : 's'}`
                                  : `Due in ${getDaysUntilDue(nextPaymentInfo.due_date)} days`}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center justify-between border-t border-border pt-3">
                        <div className="text-sm text-muted-foreground">Amount Due</div>
                        <div
                          className={`text-xl font-bold ${
                            debt.debt_type === 'to_pay' ? 'text-destructive' : 'text-success'
                          }`}
                        >
                          {formatCurrency(parseFloat(nextPaymentInfo.amount || 0))}
                        </div>
                      </div>
                      {nextPaymentInfo.payment_number && debt.number_of_payments && (
                        <div className="flex items-center justify-between border-t border-border pt-3">
                          <div className="text-sm text-muted-foreground">Payment Progress</div>
                          <div className="text-sm font-medium text-foreground">
                            #{nextPaymentInfo.payment_number} of {debt.number_of_payments}
                          </div>
                        </div>
                      )}
                      <div className="border-t border-border pt-3">
                        <button
                          onClick={() => setShowSchedule(true)}
                          className="btn-secondary w-full text-sm"
                        >
                          <svg
                            className="mr-2 h-4 w-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth="2"
                              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                            />
                          </svg>
                          View Full Payment Schedule
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div className="py-4 text-center text-sm text-muted-foreground">
                      All payments completed or no schedule available
                    </div>
                  )}
                </div>
              )}

            {/* Payment History */}
            <PaymentHistory
              debtPayments={debtPayments}
              paymentsLoading={paymentsLoading}
              showAddPayment={showAddPayment}
              setShowAddPayment={setShowAddPayment}
              newPayment={newPayment}
              setNewPayment={setNewPayment}
              receiptFile={receiptFile}
              setReceiptFile={setReceiptFile}
              isSubmitting={isSubmitting}
              onAddPayment={handleAddPayment}
              onFileChange={handleFileChange}
              onViewReceipt={(payment) => setViewingReceipt(payment)}
              debtType={debt.debt_type}
              onPaymentStatusChange={() => {
                loadPayments(debt.id)
                loadNextPaymentInfo(debt.id)
              }}
            />

            {/* Description */}
            <div className="rounded-lg border border-border p-4">
              <div className="mb-2 text-sm font-medium text-muted-foreground">Description</div>
              <div className="text-foreground">{debt.description || 'No description provided'}</div>
            </div>

            {/* Notes */}
            {debt.notes && (
              <div className="rounded-lg border border-border p-4">
                <div className="mb-2 text-sm font-medium text-muted-foreground">Notes</div>
                <div className="whitespace-pre-wrap text-foreground">{debt.notes}</div>
              </div>
            )}

            {/* Timestamps */}
            <div className="mt-6 flex flex-col justify-between space-y-2 px-1 pb-2 text-xs text-muted-foreground md:flex-row">
              <div>
                Created: <span className="text-foreground">{formatDate(debt.created_at)}</span>
              </div>
              <div>
                Last Updated: <span className="text-foreground">{formatDate(debt.updated_at)}</span>
              </div>
            </div>
          </div>
          )}

          {/* Action Buttons */}
          {activeTab === 'details' && (
          <div className="mt-6 flex space-x-3">
            <button onClick={onEdit} className="btn-primary flex-1">
              <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
              Edit
            </button>
            <button onClick={onDelete} className="btn-destructive flex-1">
              <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                />
              </svg>
              Delete
            </button>
          </div>
          )}
        </div>
      </div>

      {/* Image Viewer Modal */}
      <ImageViewerModal payment={viewingReceipt} onClose={() => setViewingReceipt(null)} />

      {/* Installment Schedule Modal */}
      {showSchedule && (
        <InstallmentScheduleModal
          debt={debt}
          payments={debtPayments}
          onClose={() => setShowSchedule(false)}
          onPaymentUpdate={() => {
            loadPayments(debt.id)
            loadNextPaymentInfo(debt.id)
          }}
        />
      )}
    </div>
  )
}
