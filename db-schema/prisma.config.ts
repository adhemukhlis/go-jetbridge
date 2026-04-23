import dotenv from 'dotenv';
import path from 'node:path';

import { defineConfig, env } from 'prisma/config';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

export default defineConfig({
	schema: 'prisma/schema.prisma',
	migrations: {
		path: 'prisma/migrations',
		seed: 'pnpm dlx tsx seed.ts',
	},
	datasource: {
		url: env('DATABASE_URL'),
	},
});
