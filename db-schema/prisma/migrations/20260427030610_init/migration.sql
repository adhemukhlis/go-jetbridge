-- CreateTable
CREATE TABLE "Credential" (
    "id" UUID NOT NULL,
    "createdAt" TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMPTZ(3) NOT NULL,
    "passwordHash" VARCHAR(255) NOT NULL,
    "suspendedAt" TIMESTAMPTZ(3),
    "suspendReason" VARCHAR(255),
    "isNeedPasswordChange" BOOLEAN NOT NULL DEFAULT true,
    "userId" UUID NOT NULL,

    CONSTRAINT "Credential_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Credential_userId_key" ON "Credential"("userId");
