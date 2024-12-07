'use client'

import React, { useState } from 'react'

interface FormFieldProps {
  label: string
  error?: string
  children: React.ReactNode
  required?: boolean
}

export function FormField({ label, error, children, required }: FormFieldProps) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {label}
        {required && <span className="text-wrale-danger ml-1">*</span>}
      </label>
      {children}
      {error && (
        <p className="mt-1 text-sm text-wrale-danger">
          {error}
        </p>
      )}
    </div>
  )
}

interface FormInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  required?: boolean
}

export function FormInput({ label, error, required, className = '', ...props }: FormInputProps) {
  return (
    <FormField label={label} error={error} required={required}>
      <input
        className={`mt-1 block w-full border-gray-300 rounded-md shadow-sm 
          focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm
          ${error ? 'border-wrale-danger' : ''} 
          ${className}`}
        {...props}
      />
    </FormField>
  )
}

interface FormSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  error?: string
  options: { value: string; label: string }[]
  required?: boolean
}

export function FormSelect({ label, error, options, required, className = '', ...props }: FormSelectProps) {
  return (
    <FormField label={label} error={error} required={required}>
      <select
        className={`mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 
          focus:outline-none focus:ring-wrale-primary focus:border-wrale-primary 
          sm:text-sm rounded-md
          ${error ? 'border-wrale-danger' : ''} 
          ${className}`}
        {...props}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </FormField>
  )
}

interface FormTextAreaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string
  error?: string
  required?: boolean
}

export function FormTextArea({ label, error, required, className = '', ...props }: FormTextAreaProps) {
  return (
    <FormField label={label} error={error} required={required}>
      <textarea
        className={`mt-1 block w-full border-gray-300 rounded-md shadow-sm 
          focus:ring-wrale-primary focus:border-wrale-primary sm:text-sm
          ${error ? 'border-wrale-danger' : ''} 
          ${className}`}
        {...props}
      />
    </FormField>
  )
}

interface FormCheckboxProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
}

export function FormCheckbox({ label, className = '', ...props }: FormCheckboxProps) {
  return (
    <label className="flex items-center">
      <input
        type="checkbox"
        className={`form-checkbox h-4 w-4 text-wrale-primary 
          focus:ring-wrale-primary border-gray-300 rounded
          ${className}`}
        {...props}
      />
      <span className="ml-2 text-sm text-gray-600">
        {label}
      </span>
    </label>
  )
}