import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: 'openapi.yaml',
  output: process.env.OSTOS_OPENAPI_OUTPUT ?? 'generated',
  plugins: ['@hey-api/client-fetch'],
});
