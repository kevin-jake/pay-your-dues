// API client for the DebtTracker backend
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

// API client class
class ApiClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`

    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    }

    // Add authorization header if token exists
    const token = localStorage.getItem('token')
    if (token) {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      }
    }

    try {
      const response = await fetch(url, config)

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      const data = await response.json()
      return data.data
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('An unexpected error occurred')
    }
  }

  // Authentication methods
  async login(credentials) {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    })

    // Transform the response to match our frontend interface
    return {
      token: response.token,
      user: {
        id: response.user.ID,
        email: response.user.Email,
        first_name: response.user.FirstName,
        last_name: response.user.LastName,
        phone: response.user.Phone,
        created_at: response.user.CreatedAt,
        updated_at: response.user.UpdatedAt,
      },
    }
  }

  async register(userData) {
    const response = await this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    })

    // Transform the response to match our frontend interface
    return {
      user: {
        id: response.user.ID,
        email: response.user.Email,
        first_name: response.user.FirstName,
        last_name: response.user.LastName,
        phone: response.user.Phone,
        created_at: response.user.CreatedAt,
        updated_at: response.user.UpdatedAt,
      },
    }
  }

  // Health check
  async healthCheck() {
    return this.request('/health')
  }

  // User Settings methods
  async getUserSettings() {
    return this.request('/settings', {
      method: 'GET',
    })
  }

  async updateUserSettings(settings) {
    return this.request('/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    })
  }

  // Contact Management methods
  async createContact(contactData) {
    const response = await this.request('/contacts', {
      method: 'POST',
      body: JSON.stringify(contactData),
    })
    console.log(response)
    return {
      id: response.id,
      name: response.name,
      email: response.email,
      phone: response.phone,
      notes: response.notes,
      created_at: response.created_at,
      updated_at: response.updated_at,
    }
  }

  async getContacts() {
    const response = await this.request('/contacts')
    console.log(response)
    return response.map((contact) => ({
      id: contact.id,
      name: contact.name,
      email: contact.email,
      phone: contact.phone,
      notes: contact.notes,
      created_at: contact.created_at,
      updated_at: contact.updated_at,
    }))
  }

  async getContact(id) {
    const response = await this.request(`/contacts/${id}`)

    return {
      id: response.id,
      name: response.name,
      email: response.email,
      phone: response.phone,
      notes: response.notes,
      created_at: response.created_at,
      updated_at: response.updated_at,
    }
  }

  async updateContact(id, contactData) {
    const response = await this.request(`/contacts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(contactData),
    })

    return {
      id: response.id,
      name: response.name,
      email: response.email,
      phone: response.phone,
      notes: response.notes,
      created_at: response.created_at,
      updated_at: response.updated_at,
    }
  }

  async deleteContact(id) {
    return this.request(`/contacts/${id}`, {
      method: 'DELETE',
    })
  }

  // Debt Management methods
  async getDebtLists() {
    const response = await this.request('/debts')

    return response.map((debt) => ({
      id: debt.id,
      contact_id: debt.contact_id,
      total_amount: debt.total_amount,
      currency: debt.currency,
      debt_type: debt.debt_type,
      installment_plan: debt.installment_plan,
      description: debt.description,
      notes: debt.notes,
      created_at: debt.created_at,
      updated_at: debt.updated_at,
      installment_amount: debt.installment_amount,
      total_payments_made: debt.total_payments_made,
      total_remaining_debt: debt.total_remaining_debt,
      status: debt.status,
      due_date: debt.due_date,
      next_payment_date: debt.next_payment_date,
      number_of_payments: debt.number_of_payments,
      contact: debt.contact
        ? {
            id: debt.contact.id,
            name: debt.contact.name,
            email: debt.contact.email,
            phone: debt.contact.phone,
            notes: debt.contact.notes,
            created_at: debt.contact.created_at,
            updated_at: debt.contact.updated_at,
          }
        : undefined,
    }))
  }

  async getDebtList(id) {
    const response = await this.request(`/debts/${id}`)

    return {
      id: response.id,
      contact_id: response.contact_id,
      total_amount: response.total_amount,
      currency: response.currency,
      debt_type: response.debt_type,
      installment_plan: response.installment_plan,
      description: response.description,
      notes: response.notes,
      created_at: response.created_at,
      updated_at: response.updated_at,
    }
  }

  async createDebtList(debtData) {
    const response = await this.request('/debts', {
      method: 'POST',
      body: JSON.stringify(debtData),
    })

    return {
      id: response.ID,
      contact_id: response.ContactID,
      total_amount: response.TotalAmount,
      currency: response.Currency,
      debt_type: response.DebtType,
      installment_plan: response.InstallmentPlan,
      description: response.Description,
      notes: response.Notes,
      created_at: response.CreatedAt,
      updated_at: response.UpdatedAt,
      installment_amount: response.InstallmentAmount,
      total_payments_made: response.TotalPaymentsMade,
      total_remaining_debt: response.TotalRemainingDebt,
      status: response.Status,
      due_date: response.DueDate,
      next_payment_date: response.NextPaymentDate,
      number_of_payments: response.NumberOfPayments,
    }
  }

  async updateDebtList(id, debtData) {
    const response = await this.request(`/debts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(debtData),
    })

    return {
      id: response.id,
      contact_id: response.contact_id,
      total_amount: response.total_amount,
      currency: response.currency,
      debt_type: response.debt_type,
      installment_plan: response.installment_plan,
      description: response.description,
      notes: response.notes,
      created_at: response.created_at,
      updated_at: response.updated_at,
    }
  }

  async deleteDebtList(id) {
    return this.request(`/debts/${id}`, {
      method: 'DELETE',
    })
  }

  // Payment Management methods
  async createPayment(debtId, paymentData) {
    const response = await this.request('/debts/payments', {
      method: 'POST',
      body: JSON.stringify({
        debt_list_id: debtId,
        amount: paymentData.amount,
        payment_date: paymentData.payment_date,
        payment_method: paymentData.payment_method || 'cash',
        description: paymentData.description,
      }),
    })

    return this._mapPayment(response)
  }

  // Helper method to map payment response
  _mapPayment(payment) {
    return {
      id: payment.ID,
      debt_list_id: payment.DebtListID,
      amount: payment.Amount,
      currency: payment.Currency,
      payment_date: payment.PaymentDate,
      payment_method: payment.PaymentMethod,
      description: payment.Description,
      status: payment.Status,
      receipt_photo_url: payment.ReceiptPhotoURL,
      verified_by: payment.VerifiedBy,
      verified_at: payment.VerifiedAt,
      verification_notes: payment.VerificationNotes,
      created_at: payment.CreatedAt,
      updated_at: payment.UpdatedAt,
    }
  }

  async getPayments(debtId) {
    const response = await this.request(`/debts/${debtId}/payments`)
    return response.map((payment) => this._mapPayment(payment))
  }

  async updatePayment(paymentId, paymentData) {
    const response = await this.request(`/debts/payments/${paymentId}`, {
      method: 'PUT',
      body: JSON.stringify(paymentData),
    })

    return this._mapPayment(response)
  }

  async deletePayment(paymentId) {
    return this.request(`/debts/payments/${paymentId}`, {
      method: 'DELETE',
    })
  }

  async uploadReceiptPhoto(paymentId, file) {
    const formData = new FormData()
    formData.append('receipt', file)

    const token = localStorage.getItem('token')
    const config = {
      method: 'POST',
      body: formData,
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    }

    const response = await fetch(`${this.baseUrl}/debts/payments/${paymentId}/receipt`, config)

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || `HTTP ${response.status}`)
    }

    const data = await response.json()
    return {
      receipt_url: data.data.receipt_url,
    }
  }

  async getUpcomingPayments(days = 30) {
    const response = await this.request(`/upcoming-payments?days=${days}`)
    return response
  }

  async getPaymentSchedule(debtId) {
    const response = await this.request(`/debts/${debtId}/schedule`)
    return response
  }

  async verifyPayment(paymentId) {
    const response = await this.request(`/debts/payments/${paymentId}/verify`, {
      method: 'POST',
      body: JSON.stringify({
        status: 'completed',
      }),
    })
    return this._mapPayment(response)
  }

  async rejectPayment(paymentId) {
    const response = await this.request(`/debts/payments/${paymentId}/reject`, {
      method: 'POST',
      body: JSON.stringify({
        status: 'rejected',
      }),
    })
    return this._mapPayment(response)
  }

  // Method to fetch images with authorization headers
  async fetchImageWithAuth(imageUrl) {
    const token = localStorage.getItem('token')

    const headers = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    try {
      const response = await fetch(imageUrl, {
        method: 'GET',
        headers,
      })

      if (!response.ok) {
        throw new Error(`Failed to fetch image: ${response.status}`)
      }

      const blob = await response.blob()
      return URL.createObjectURL(blob)
    } catch (error) {
      throw error
    }
  }
}

// Create and export the API client instance
export const apiClient = new ApiClient(API_BASE_URL)

// Utility functions for token and user data management
export const tokenManager = {
  getToken() {
    return localStorage.getItem('token')
  },

  setToken(token) {
    localStorage.setItem('token', token)
  },

  removeToken() {
    localStorage.removeItem('token')
    this.removeUserData() // Also remove user data when removing token
  },

  hasToken() {
    return !!this.getToken()
  },

  getUserData() {
    const userData = localStorage.getItem('user')
    if (userData) {
      try {
        return JSON.parse(userData)
      } catch (error) {
        console.error('Failed to parse user data from localStorage:', error)
        return null
      }
    }
    return null
  },

  setUserData(user) {
    if (user) {
      localStorage.setItem('user', JSON.stringify(user))
    }
  },

  removeUserData() {
    localStorage.removeItem('user')
  },
}
