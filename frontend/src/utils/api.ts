// Use the environment variable if it's available, otherwise fall back to the default for local development.
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

// The base URL for the backend server, used for constructing image URLs.
// It's derived by removing the '/api' suffix from the API_BASE_URL.
export const BACKEND_BASE_URL = API_BASE_URL.replace('/api', '');

