package config

// Etcd server hostname
const ETCD_ADDRESS = "etcd.address"

// exposed port for serverledge APIs
const API_PORT = "api.port"
const API_IP = "api.ip"

// REMOTE SERVER URL
const CLOUD_URL = "cloud.server.url"

// Forces runtime container images to be pulled the first time they are used,
// even if they are locally available (true/false).
const FACTORY_REFRESH_IMAGES = "factory.images.refresh"

// Amount of memory available for the container pool (in MB)
const POOL_MEMORY_MB = "container.pool.memory"

// CPUs available for the container pool (1.0 = 1 core)
const POOL_CPUS = "container.pool.cpus"

// periodically janitor wakes up and deletes expired containers
const POOL_CLEANUP_PERIOD = "janitor.interval"

// container expiration time
const CONTAINER_EXPIRATION_TIME = "container.expiration"

// cache capacity
const CACHE_SIZE = "cache.size"

// cache janitor interval (Seconds) : deletes expired items
const CACHE_CLEANUP = "cache.cleanup"

// default expiration time assigned to a cache item (Seconds)
const CACHE_ITEM_EXPIRATION = "cache.expiration"

// true if the current server is a remote cloud server
const IS_IN_CLOUD = "cloud"

// the area wich the server belongs to
const REGISTRY_AREA = "registry.area"

// short period: retrieve information about nearby edge-servers
const REG_NEARBY_INTERVAL = "registry.nearby.interval"

// long period for general monitoring inside the area
const REG_MONITORING_INTERVAL = "registry.monitoring.interval"

// registration TTL in seconds
const REGISTRATION_TTL = "registry.ttl"

// port for udp status listener
const LISTEN_UDP_PORT = "registry.udp.port"

// enable metrics system
const METRICS_ENABLED = "metrics.enabled"

const METRICS_PORT = "metrics.port"

// Bandwidth between edge-cloud
const BANDWIDTH_CLOUD = "metrics.bandwidth.cloud"

// Badnwidth between edge-edge
const BANDWIDTH_EDGE = "metrics.bandwidth.cloud"

// Scheduling policy to use
// Possible values: "qosaware", "default", "cloudonly"
const SCHEDULING_POLICY = "scheduler.policy"

const CLOUD_COST_FACTOR = "scheduler.cloud.cost"
const BUDGET = "scheduler.local.budget"

// Capacity of the queue (possibly) used by the scheduler
const SCHEDULER_QUEUE_CAPACITY = "scheduler.queue.capacity"

// Solver interval
const SOLVER_EVALUATION_INTERVAL = "solver.evalinterval"

// Solver ip address
const SOLVER_ADDRESS = "solver.address"

const STORAGE_VERSION = "storage.version"

// InfluxDB
const STORAGE_DB_ADDRESS = "storage.address"
const STORAGE_DB_TOKEN = "storage.token"
const STORAGE_DB_ORGNAME = "storage.orgname"

const DOCKER_LIMIT_CPU = "docker.cpu"

// Cost of the cloud node
const CLOUD_NODE_COST = "cloud.cost" // Cost of the cloud node (float)
const CLOUD_DELAY = "cloud.delay"    // Delay of execution

// MAB agent
const MAB_AGENT_ENABLED = "mab.agent.enabled"                       // True if MAB agent is enabled (bool)
const MAB_AGENT_INTERVAL = "mab.agent.interval"                     // Update interval of MAB agent in seconds (int)
const MAB_AGENT_STRATEGY = "mab.agent.strategy"                     // Exploration strategy of the MAB agent (string: "Epsilon-Greedy", "UCB", "ResetUCB", "SWUCB")
const MAB_AGENT_REWARD_ALPHA = "mab.agent.reward.alpha"             // Coefficient for load imbalance (float)
const MAB_AGENT_REWARD_BETA = "mab.agent.reward.beta"               // Coefficient for response time (float)
const MAB_AGENT_REWARD_GAMMA = "mab.agent.reward.gamma"             // Coefficient for cost (float)
const MAB_AGENT_REWARD_DELTA = "mab.agent.reward.delta"             // Coefficient for utility (float)
const MAB_AGENT_REWARD_ZETA = "mab.agent.reward.zeta"               // Coefficient for violations count (float)
const MAB_AGENT_EPSILON = "mab.agent.epsilon"                       // Probability of exploration for the epsilon-greedy strategy (float)
const MAB_AGENT_EXPLORATIONFACTOR = "mab.agent.explorationfactor"   // Exploration factor for UCB, ResetUCB and SWUCB strategies (float)
const MAB_AGENT_SWUCB_WINDOWSIZE = "mab.agent.swucb.windowsize"     // Size of the sliding window for SWUCB strategy (int)
const MAB_AGENT_RUCB_RESETINTERVAL = "mab.agent.rucb.resetinterval" // Reset interval for ResetUCB strategy (int)
const MAB_AGENT_UCB2_ALPHA = "mab.agent.ucb2.alpha"                 // \alpha parameter for UCB2
const MAB_AGENT_KLUCB_C = "mab.agent.klucb.c"                       // \alpha parameter for UCB2
