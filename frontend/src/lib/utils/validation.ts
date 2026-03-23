// Mirrors backend constraints from svc-gateway/internal/dto/user.go
// and svc-gateway/pkg/validator/validator.go

// ^[a-zA-Z0-9_]+$
const USERNAME_REGEX = /^[a-zA-Z0-9_]+$/;

// ^[a-zA-Z0-9_!@#$%^&*]+$
const PASSWORD_REGEX = /^[a-zA-Z0-9_!@#$%^&*]+$/;

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// Returns an i18n key string on error, null on success
export function validateUsername(value: string): string | null {
	if (!value) return 'validation.username.required';
	if (value.length < 3) return 'validation.username.min';
	if (value.length > 20) return 'validation.username.max';
	if (!USERNAME_REGEX.test(value)) return 'validation.username.invalid';
	return null;
}

export function validatePassword(value: string): string | null {
	if (!value) return 'validation.password.required';
	if (value.length < 6) return 'validation.password.min';
	if (value.length > 50) return 'validation.password.max';
	if (!PASSWORD_REGEX.test(value)) return 'validation.password.invalid';
	return null;
}

export function validateEmail(value: string): string | null {
	if (!value) return 'validation.email.required';
	if (!EMAIL_REGEX.test(value)) return 'validation.email.invalid';
	return null;
}

export function validateEmailCode(value: string): string | null {
	if (!value) return 'validation.email_code.required';
	if (!/^\d{6}$/.test(value)) return 'validation.email_code.invalid';
	return null;
}

// Validates login identifier: email or username depending on presence of '@'
export function validateIdentifier(value: string): string | null {
	if (!value) return 'validation.identifier.required';
	return value.includes('@') ? validateEmail(value) : validateUsername(value);
}

export interface LoginFormErrors {
	identifier?: string;
	password?: string;
}

export interface RegisterFormErrors {
	username?: string;
	email?: string;
	password?: string;
	confirmed_password?: string;
	email_code?: string;
}

export interface ResetFormErrors {
	email?: string;
	email_code?: string;
	password?: string;
	confirmed_password?: string;
}

export function validateLoginForm(form: {
	identifier: string;
	password: string;
}): LoginFormErrors | null {
	const errors: LoginFormErrors = {};
	const identifierErr = validateIdentifier(form.identifier);
	if (identifierErr) errors.identifier = identifierErr;
	const passwordErr = validatePassword(form.password);
	if (passwordErr) errors.password = passwordErr;
	return Object.keys(errors).length ? errors : null;
}

export function validateRegisterForm(form: {
	username: string;
	email: string;
	password: string;
	confirmed_password: string;
	email_code: string;
}): RegisterFormErrors | null {
	const errors: RegisterFormErrors = {};
	const usernameErr = validateUsername(form.username);
	if (usernameErr) errors.username = usernameErr;
	const emailErr = validateEmail(form.email);
	if (emailErr) errors.email = emailErr;
	const passwordErr = validatePassword(form.password);
	if (passwordErr) errors.password = passwordErr;
	if (!form.confirmed_password) {
		errors.confirmed_password = 'validation.confirmed_password.required';
	} else if (form.confirmed_password !== form.password) {
		errors.confirmed_password = 'validation.confirmed_password.mismatch';
	}
	const codeErr = validateEmailCode(form.email_code);
	if (codeErr) errors.email_code = codeErr;
	return Object.keys(errors).length ? errors : null;
}

export function validateResetForm(form: {
	email: string;
	email_code: string;
	password: string;
	confirmed_password: string;
}): ResetFormErrors | null {
	const errors: ResetFormErrors = {};
	const emailErr = validateEmail(form.email);
	if (emailErr) errors.email = emailErr;
	const codeErr = validateEmailCode(form.email_code);
	if (codeErr) errors.email_code = codeErr;
	const passwordErr = validatePassword(form.password);
	if (passwordErr) errors.password = passwordErr;
	if (!form.confirmed_password) {
		errors.confirmed_password = 'validation.confirmed_password.required';
	} else if (form.confirmed_password !== form.password) {
		errors.confirmed_password = 'validation.confirmed_password.mismatch';
	}
	return Object.keys(errors).length ? errors : null;
}
