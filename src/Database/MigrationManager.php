<?php

declare(strict_types=1);

namespace Dolphin\Database;

use Dolphin\Core\Application;
use Dolphin\Database\Schema\SchemaBuilder;
use Dolphin\Database\Schema\Blueprint;
use PDO;
use PDOException;

/**
 * Migration Interface
 * 
 * Inspired by Raptor's migration system for Go
 * Each migration implements Up() and Down() methods
 */
interface Migration
{
    /**
     * Get the migration name
     */
    public function name(): string;

    /**
     * Run the migration
     */
    public function up(SchemaBuilder $schema): void;

    /**
     * Reverse the migration
     */
    public function down(SchemaBuilder $schema): void;
}

/**
 * Migration Manager
 * 
 * Manages database migrations with batch execution and rollback capabilities
 * Similar to Raptor's migration management system
 */
class MigrationManager
{
    protected PDO $connection;
    protected string $migrationsPath;
    protected string $tableName = 'migrations';

    public function __construct(PDO $connection, string $migrationsPath)
    {
        $this->connection = $connection;
        $this->migrationsPath = $migrationsPath;
        $this->ensureMigrationsTable();
    }

    /**
     * Ensure the migrations table exists
     */
    protected function ensureMigrationsTable(): void
    {
        $sql = "CREATE TABLE IF NOT EXISTS {$this->tableName} (
            id INT AUTO_INCREMENT PRIMARY KEY,
            migration VARCHAR(255) NOT NULL,
            batch INT NOT NULL,
            executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE KEY unique_migration (migration)
        )";
        
        $this->connection->exec($sql);
    }

    /**
     * Run all pending migrations
     */
    public function migrate(): array
    {
        $pendingMigrations = $this->getPendingMigrations();
        
        if (empty($pendingMigrations)) {
            return ['message' => 'No pending migrations'];
        }

        $batch = $this->getNextBatchNumber();
        $executed = [];

        foreach ($pendingMigrations as $migration) {
            try {
                $this->connection->beginTransaction();
                
                $instance = new $migration();
                $schema = new SchemaBuilder($this->connection);
                
                $instance->up($schema);
                
                $this->recordMigration($migration, $batch);
                $this->connection->commit();
                
                $executed[] = $migration;
            } catch (Exception $e) {
                $this->connection->rollBack();
                throw new Exception("Migration failed: {$migration} - " . $e->getMessage());
            }
        }

        return [
            'message' => 'Migrations completed successfully',
            'executed' => $executed,
            'batch' => $batch
        ];
    }

    /**
     * Rollback the last batch of migrations
     */
    public function rollback(): array
    {
        $lastBatch = $this->getLastBatchNumber();
        
        if (!$lastBatch) {
            return ['message' => 'No migrations to rollback'];
        }

        $migrations = $this->getMigrationsByBatch($lastBatch);
        $rolledBack = [];

        foreach (array_reverse($migrations) as $migration) {
            try {
                $this->connection->beginTransaction();
                
                $instance = new $migration();
                $schema = new SchemaBuilder($this->connection);
                
                $instance->down($schema);
                
                $this->removeMigration($migration);
                $this->connection->commit();
                
                $rolledBack[] = $migration;
            } catch (Exception $e) {
                $this->connection->rollBack();
                throw new Exception("Rollback failed: {$migration} - " . $e->getMessage());
            }
        }

        return [
            'message' => 'Rollback completed successfully',
            'rolled_back' => $rolledBack,
            'batch' => $lastBatch
        ];
    }

    /**
     * Get migration status
     */
    public function status(): array
    {
        $allMigrations = $this->getAllMigrations();
        $executedMigrations = $this->getExecutedMigrations();
        
        $status = [];
        
        foreach ($allMigrations as $migration) {
            $status[] = [
                'migration' => $migration,
                'status' => in_array($migration, $executedMigrations) ? 'executed' : 'pending',
                'batch' => $this->getMigrationBatch($migration)
            ];
        }

        return $status;
    }

    /**
     * Get all migration files
     */
    protected function getAllMigrations(): array
    {
        $files = glob($this->migrationsPath . '/*.php');
        $migrations = [];

        foreach ($files as $file) {
            $className = basename($file, '.php');
            $migrations[] = $className;
        }

        sort($migrations);
        return $migrations;
    }

    /**
     * Get pending migrations
     */
    protected function getPendingMigrations(): array
    {
        $allMigrations = $this->getAllMigrations();
        $executedMigrations = $this->getExecutedMigrations();
        
        return array_diff($allMigrations, $executedMigrations);
    }

    /**
     * Get executed migrations
     */
    protected function getExecutedMigrations(): array
    {
        $stmt = $this->connection->query("SELECT migration FROM {$this->tableName}");
        return $stmt->fetchAll(PDO::FETCH_COLUMN);
    }

    /**
     * Get migrations by batch number
     */
    protected function getMigrationsByBatch(int $batch): array
    {
        $stmt = $this->connection->prepare("SELECT migration FROM {$this->tableName} WHERE batch = ?");
        $stmt->execute([$batch]);
        return $stmt->fetchAll(PDO::FETCH_COLUMN);
    }

    /**
     * Get the next batch number
     */
    protected function getNextBatchNumber(): int
    {
        $stmt = $this->connection->query("SELECT MAX(batch) as max_batch FROM {$this->tableName}");
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        return ($result['max_batch'] ?? 0) + 1;
    }

    /**
     * Get the last batch number
     */
    protected function getLastBatchNumber(): ?int
    {
        $stmt = $this->connection->query("SELECT MAX(batch) as max_batch FROM {$this->tableName}");
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        return $result['max_batch'] ? (int)$result['max_batch'] : null;
    }

    /**
     * Get migration batch number
     */
    protected function getMigrationBatch(string $migration): ?int
    {
        $stmt = $this->connection->prepare("SELECT batch FROM {$this->tableName} WHERE migration = ?");
        $stmt->execute([$migration]);
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        return $result ? (int)$result['batch'] : null;
    }

    /**
     * Record a migration as executed
     */
    protected function recordMigration(string $migration, int $batch): void
    {
        $stmt = $this->connection->prepare(
            "INSERT INTO {$this->tableName} (migration, batch) VALUES (?, ?)"
        );
        $stmt->execute([$migration, $batch]);
    }

    /**
     * Remove a migration record
     */
    protected function removeMigration(string $migration): void
    {
        $stmt = $this->connection->prepare("DELETE FROM {$this->tableName} WHERE migration = ?");
        $stmt->execute([$migration]);
    }
}
