import dotenv from 'dotenv';
import path from 'node:path';
import { PrismaPg } from '@prisma/adapter-pg';

import { Prisma, PrismaClient } from './generated/prisma/client';
import { PrismaTransactionType } from './types';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

const DATABASE_URL = `postgresql://${process.env.PGUSER}:${process.env.PGPASSWORD}@${process.env.PGHOST}:${process.env.PGPORT}/${process.env.PGDATABASE}?sslmode=${process.env.SSL_MODE}`;

const adapter = new PrismaPg({
	connectionString: DATABASE_URL,
	ssl: process.env.SSL_MODE === 'require' ? { rejectUnauthorized: false } : false,
});

const prismaClient = new PrismaClient({
	transactionOptions: {
		isolationLevel: Prisma.TransactionIsolationLevel.Serializable,
		maxWait: 1 * 1000 * 60 * 60,
		timeout: 3 * 1000 * 60 * 60,
	},
	adapter,
});

main();

async function main() {
	try {
		await prismaClient.$transaction(async (prisma: PrismaTransactionType) => {
			await prisma.user.createMany({
				data: {
					name: 'Admin',
					username: 'admin',
					email: 'admin@mail.com',
				},
				skipDuplicates: true,
			});

			await prisma.role.createMany({
				data: {
					name: 'Super Admin',
					key: 'super_admin',
				},
				skipDuplicates: true,
			});
		});
	} catch (error) {
		console.error('Script failed ❌', error);
		process.exit(1);
	} finally {
		await prismaClient.$disconnect();
	}
}
