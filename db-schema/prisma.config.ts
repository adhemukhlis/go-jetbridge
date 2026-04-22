import dotenv from 'dotenv';
import path from 'node:path';

import { defineConfig, env } from 'prisma/config';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

const DATABASE_URL = `postgresql://${process.env.PGUSER}:${process.env.PGPASSWORD}@${process.env.PGHOST}:${process.env.PGPORT}/${process.env.PGDATABASE}?sslmode=${process.env.SSL_MODE}`;

export default defineConfig({
	schema: 'prisma/schema.prisma',
	migrations: {
		path: 'prisma/migrations',
		seed: 'pnpm dlx tsx seed.ts',
	},
	datasource: {
		url: DATABASE_URL,
	},
});
