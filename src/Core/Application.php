<?php

declare(strict_types=1);

namespace Dolphin\Core;

use Dolphin\Container\Container;
use Dolphin\Http\Kernel;
use Dolphin\Routing\Router;
use Dolphin\Config\Config;
use Dolphin\Database\DatabaseManager;
use Dolphin\View\ViewManager;
use Dolphin\Cache\CacheManager;
use Dolphin\Session\SessionManager;
use Dolphin\Log\Logger;
use Dotenv\Dotenv;

/**
 * Application Class
 * 
 * The main application class that bootstraps the framework
 * and manages the application lifecycle.
 */
class Application
{
    protected Container $container;
    protected string $basePath;
    protected string $environment;
    protected bool $booted = false;

    public function __construct(string $basePath = null)
    {
        $this->basePath = $basePath ?: dirname(__DIR__, 2);
        $this->container = new Container();
        $this->environment = $this->detectEnvironment();
        
        $this->registerBaseBindings();
        $this->loadEnvironment();
        $this->registerCoreServices();
    }

    /**
     * Detect the application environment
     */
    protected function detectEnvironment(): string
    {
        return $_ENV['APP_ENV'] ?? $_SERVER['APP_ENV'] ?? 'production';
    }

    /**
     * Register base bindings in the container
     */
    protected function registerBaseBindings(): void
    {
        $this->container->instance('app', $this);
        $this->container->instance(Application::class, $this);
    }

    /**
     * Load environment variables
     */
    protected function loadEnvironment(): void
    {
        $dotenv = Dotenv::createImmutable($this->basePath);
        $dotenv->safeLoad();
    }

    /**
     * Register core services
     */
    protected function registerCoreServices(): void
    {
        // Configuration
        $this->container->singleton('config', function () {
            return new Config($this->basePath);
        });

        // Logger
        $this->container->singleton('log', function () {
            return new Logger($this->container->get('config'));
        });

        // Database Manager
        $this->container->singleton('db', function () {
            return new DatabaseManager($this->container->get('config'));
        });

        // Router
        $this->container->singleton('router', function () {
            return new Router();
        });

        // View Manager
        $this->container->singleton('view', function () {
            return new ViewManager($this->container->get('config'));
        });

        // Cache Manager
        $this->container->singleton('cache', function () {
            return new CacheManager($this->container->get('config'));
        });

        // Session Manager
        $this->container->singleton('session', function () {
            return new SessionManager($this->container->get('config'));
        });

        // HTTP Kernel
        $this->container->singleton(Kernel::class, function () {
            return new Kernel($this);
        });
    }

    /**
     * Boot the application
     */
    public function boot(): void
    {
        if ($this->booted) {
            return;
        }

        $this->booted = true;
        
        // Load service providers
        $this->loadServiceProviders();
        
        // Boot service providers
        $this->bootServiceProviders();
    }

    /**
     * Load service providers
     */
    protected function loadServiceProviders(): void
    {
        $providers = $this->container->get('config')->get('app.providers', []);
        
        foreach ($providers as $provider) {
            $this->container->register($provider);
        }
    }

    /**
     * Boot service providers
     */
    protected function bootServiceProviders(): void
    {
        $providers = $this->container->get('config')->get('app.providers', []);
        
        foreach ($providers as $provider) {
            $instance = $this->container->get($provider);
            if (method_exists($instance, 'boot')) {
                $instance->boot();
            }
        }
    }

    /**
     * Get the container instance
     */
    public function getContainer(): Container
    {
        return $this->container;
    }

    /**
     * Get a service from the container
     */
    public function make(string $abstract, array $parameters = [])
    {
        return $this->container->get($abstract, $parameters);
    }

    /**
     * Get the base path
     */
    public function basePath(string $path = ''): string
    {
        return $this->basePath . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : '');
    }

    /**
     * Get the application path
     */
    public function appPath(string $path = ''): string
    {
        return $this->basePath('app' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the config path
     */
    public function configPath(string $path = ''): string
    {
        return $this->basePath('config' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the database path
     */
    public function databasePath(string $path = ''): string
    {
        return $this->basePath('database' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the public path
     */
    public function publicPath(string $path = ''): string
    {
        return $this->basePath('public' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the resources path
     */
    public function resourcePath(string $path = ''): string
    {
        return $this->basePath('resources' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the storage path
     */
    public function storagePath(string $path = ''): string
    {
        return $this->basePath('storage' . ($path ? DIRECTORY_SEPARATOR . ltrim($path, DIRECTORY_SEPARATOR) : ''));
    }

    /**
     * Get the environment
     */
    public function environment(): string
    {
        return $this->environment;
    }

    /**
     * Check if the application is in production
     */
    public function isProduction(): bool
    {
        return $this->environment === 'production';
    }

    /**
     * Check if the application is in development
     */
    public function isDevelopment(): bool
    {
        return $this->environment === 'development' || $this->environment === 'local';
    }

    /**
     * Check if the application is in testing
     */
    public function isTesting(): bool
    {
        return $this->environment === 'testing';
    }

    /**
     * Check if debugging is enabled
     */
    public function isDebugEnabled(): bool
    {
        return $this->container->get('config')->get('app.debug', false);
    }
}
