#include "ispappcwmp.h"
#include <net/if.h>
#define IFNAMSIZ 16
static void json_handle_check_parameter_value_change(void);
static void ISPAPPCWMP_do_reload(struct uloop_timeout *timeout);
static void ISPAPPCWMP_do_notify(struct uloop_timeout *timeout);
static void netlink_new_msg(struct uloop_fd *ufd, unsigned events);

static struct uloop_fd netlink_event = { .cb = netlink_new_msg };
static struct uloop_timeout reload_timer = { .cb = ISPAPPCWMP_do_reload };
static struct uloop_timeout notify_timer = { .cb = ISPAPPCWMP_do_notify };

struct option long_opts[] = {
	{"foreground", no_argument, NULL, 'f'},
	{"help", no_argument, NULL, 'h'},
	{"version", no_argument, NULL, 'v'},
	{"boot", no_argument, NULL, 'b'},
	{"getrpcmethod", no_argument, NULL, 'g'},
	{NULL, 0, NULL, 0}
};

static void print_help(void)
{
	printf("Usage: %s [OPTIONS]\n", NAME);
	printf(" -f, --foreground        Run in the foreground\n");
	printf(" -b, --boot              Run with \"1 BOOT\" event\n");
	printf(" -g, --getrpcmethod      Run with \"2 PERIODIC\" event and with ACS GetRPCMethods\n");
	printf(" -h, --help              Display this help text\n");
	printf(" -v, --version           Display the %s version\n", NAME);
}

static void print_version(void)
{
	printf("%s version: %s\n", NAME, ISPAPPCWMP_VERSION);
}

static void ISPAPPCWMP_do_reload(struct uloop_timeout *timeout)
{
	log_message(NAME, L_NOTICE, "configuration reload\n");
	if (external_init()) {
		log_message(NAME, L_DEBUG, "external scripts initialization failed\n");
		return;
	}
	config_load();
	external_exit();
}

static void ISPAPPCWMP_do_notify(struct uloop_timeout *timeout)
{
	log_message(NAME, L_NOTICE, "checking if there is notify value change\n");
	if (external_init()) {
		log_message(NAME, L_DEBUG, "external scripts initialization failed\n");
		return;
	}
	external_action_simple_execute("check_value_change", NULL, NULL);
	external_action_handle(json_handle_check_parameter_value_change);
	external_exit();
}

void ISPAPPCWMP_reload(void)
{
	uloop_timeout_set(&reload_timer, 100);
}

void ISPAPPCWMP_notify(void)
{
	uloop_timeout_set(&notify_timer, 1);
}


static void ISPAPPCWMP_netlink_interface(struct nlmsghdr *nlh)
{
	struct ifaddrmsg *ifa = (struct ifaddrmsg *) NLMSG_DATA(nlh);
	struct rtattr *rth = IFA_RTA(ifa);
	int rtl = IFA_PAYLOAD(nlh);
	char if_name[IFNAMSIZ], if_addr[INET_ADDRSTRLEN];
	static uint32_t old_addr=0;

	memset(&if_name, 0, sizeof(if_name));
	memset(&if_addr, 0, sizeof(if_addr));

	while (rtl && RTA_OK(rth, rtl)) {
		if (rth->rta_type != IFA_LOCAL) {
			rth = RTA_NEXT(rth, rtl);
			continue;
		}

		uint32_t addr = htonl(* (uint32_t *)RTA_DATA(rth));
		if (htonl(13) == 13) {
			// running on big endian system
		} else {
			// running on little endian system
			addr = __builtin_bswap32(addr);
		}

		if_indextoname(ifa->ifa_index, if_name);
		if (strncmp(config->local->interface, if_name, IFNAMSIZ)) {
			rth = RTA_NEXT(rth, rtl);
			continue;
		}

		if ((addr != old_addr) && (old_addr != 0)) {
			log_message(NAME, L_NOTICE, "ip address of the interface %s is changed\n",	if_name);
			cwmp_add_event(EVENT_VALUE_CHANGE, NULL, 0, EVENT_NO_BACKUP);
			cwmp_add_inform_timer();
		}
		old_addr = addr;

		inet_ntop(AF_INET, &(addr), if_addr, INET_ADDRSTRLEN);

		if (config->local) FREE(config->local->ip);
		config->local->ip = strdup(if_addr);
		break;
	}

	if (strlen(if_addr) == 0) return;

	log_message(NAME, L_NOTICE, "interface %s has ip %s\n",
			if_name, if_addr);
}

static void
netlink_new_msg(struct uloop_fd *ufd, unsigned events)
{
	struct nlmsghdr *nlh;
	char buffer[BUFSIZ];
	int msg_size;

	memset(&buffer, 0, sizeof(buffer));

	nlh = (struct nlmsghdr *)buffer;
	if ((msg_size = recv(ufd->fd, nlh, BUFSIZ, 0)) == -1) {
		log_message(NAME, L_DEBUG, "error receiving netlink message\n");
		return;
	}

	while (msg_size > sizeof(*nlh)) {
		int len = nlh->nlmsg_len;
		int req_len = len - sizeof(*nlh);

		if (req_len < 0 || len > msg_size) {
			log_message(NAME, L_DEBUG, "error reading netlink message\n");
			return;
		}

		if (!NLMSG_OK(nlh, msg_size)) {
			log_message(NAME, L_DEBUG, "netlink message is not NLMSG_OK\n");
			return;
		}

		if (nlh->nlmsg_type == RTM_NEWADDR)
			ISPAPPCWMP_netlink_interface(nlh);

		msg_size -= NLMSG_ALIGN(len);
		nlh = (struct nlmsghdr*)((char*)nlh + NLMSG_ALIGN(len));
	}
}

static int netlink_init(void)
{
	struct {
		struct nlmsghdr hdr;
		struct ifaddrmsg msg;
	} req;
	struct sockaddr_nl addr;

	memset(&addr, 0, sizeof(addr));
	memset(&req, 0, sizeof(req));

	if ((cwmp->netlink_sock[0] = socket(PF_NETLINK, SOCK_RAW, NETLINK_ROUTE)) == -1) {
		log_message(NAME, L_DEBUG, "couldn't open NETLINK_ROUTE socket");
		return -1;
	}
	if (fcntl(cwmp->netlink_sock[0], F_SETFD, fcntl(cwmp->netlink_sock[0], F_GETFD) | FD_CLOEXEC) < 0)
		log_message(NAME, L_NOTICE, "error in fcntl\n");

	addr.nl_family = AF_NETLINK;
	addr.nl_groups = RTMGRP_IPV4_IFADDR;
	if ((bind(cwmp->netlink_sock[0], (struct sockaddr *)&addr, sizeof(addr))) == -1) {
		log_message(NAME, L_DEBUG, "couldn't bind netlink socket");
		return -1;
	}

	netlink_event.fd = cwmp->netlink_sock[0];
	uloop_fd_add(&netlink_event, ULOOP_READ | ULOOP_EDGE_TRIGGER);

	if ((cwmp->netlink_sock[1] = socket(PF_NETLINK, SOCK_DGRAM, NETLINK_ROUTE)) == -1) {
		log_message(NAME, L_DEBUG, "couldn't open NETLINK_ROUTE socket");
		return -1;
	}
	if (fcntl(cwmp->netlink_sock[1], F_SETFD, fcntl(cwmp->netlink_sock[1], F_GETFD) | FD_CLOEXEC) < 0)
		log_message(NAME, L_NOTICE, "error in fcntl\n");

	req.hdr.nlmsg_len = NLMSG_LENGTH(sizeof(struct ifaddrmsg));
	req.hdr.nlmsg_flags = NLM_F_REQUEST | NLM_F_ROOT;
	req.hdr.nlmsg_type = RTM_GETADDR;
	req.msg.ifa_family = AF_INET;

	if ((send(cwmp->netlink_sock[1], &req, req.hdr.nlmsg_len, 0)) == -1) {
		log_message(NAME, L_DEBUG, "couldn't send netlink socket");
		return -1;
	}

	struct uloop_fd dummy_event = { .fd = cwmp->netlink_sock[1] };
	netlink_new_msg(&dummy_event, 0);

	return 0;
}

int main (int argc, char **argv)
{
	int c;
	int start_event = 0;
	bool foreground = false;

	while (1) {
		c = getopt_long(argc, argv, "fhbgv", long_opts, NULL);
		if (c == EOF)
			break;
		switch (c) {
			case 'b':
				start_event |= START_BOOT;
				break;
			case 'f':
				foreground = true;
				break;
			case 'g':
				start_event |= START_GET_RPC_METHOD;
				break;
			case 'h':
				print_help();
				exit(EXIT_SUCCESS);
			case 'v':
				print_version();
				exit(EXIT_SUCCESS);
			default:
				print_help();
				exit(EXIT_FAILURE);
		}
	}
	int fd = open("/var/run/ispappcwmp.pid", O_RDWR | O_CREAT, S_IRUSR | S_IWUSR | S_IRGRP | S_IROTH);
	if(fd == -1)
		exit(EXIT_FAILURE);
	if (flock(fd, LOCK_EX | LOCK_NB) == -1)
		exit(EXIT_SUCCESS);
	if(fcntl(fd, F_SETFD, fcntl(fd, F_GETFD) | FD_CLOEXEC) < 0 )
		log_message(NAME, L_NOTICE, "error in fcntl\n");

	setlocale(LC_CTYPE, "");
	umask(0037);

	if (getuid() != 0) {
		log_message(NAME, L_DEBUG, "run %s as root\n", NAME);
		exit(EXIT_FAILURE);
	}

	/* run early cwmp initialization */
	cwmp = calloc(1, sizeof(struct cwmp_internal));
	if (!cwmp) return -1;

	INIT_LIST_HEAD(&cwmp->events);
	INIT_LIST_HEAD(&cwmp->notifications);
	INIT_LIST_HEAD(&cwmp->downloads);
	INIT_LIST_HEAD(&cwmp->uploads);
	INIT_LIST_HEAD(&cwmp->scheduled_informs);
	uloop_init();
	backup_init();
	if (external_init()) {
		log_message(NAME, L_DEBUG, "external scripts initialization failed\n");
		return -1;
	}
	config_load();
	log_message(NAME, L_NOTICE, "daemon started\n");
	cwmp_init_deviceid();

	external_exit();

	if (start_event & START_BOOT) {
		cwmp_add_event(EVENT_BOOT, NULL, 0, EVENT_BACKUP);
		cwmp_add_inform_timer();
	}
	if (start_event & START_GET_RPC_METHOD) {
		cwmp->get_rpc_methods = true;
		cwmp_add_event(EVENT_PERIODIC, NULL, 0, EVENT_BACKUP);
		cwmp_add_inform_timer();
	}

	if (netlink_init()) {
		log_message(NAME, L_DEBUG, "netlink initialization failed\n");
		exit(EXIT_FAILURE);
	}

	if (ubus_init()) log_message(NAME, L_DEBUG, "ubus initialization failed\n");

	// Initialize command execution module
	if (command_init()) {
		log_message(NAME, L_DEBUG, "command module initialization failed\n");
		exit(EXIT_FAILURE);
	}

	http_server_init();

	pid_t pid, sid;
	if (!foreground) {
		pid = fork();
		if (pid < 0)
			exit(EXIT_FAILURE);
		if (pid > 0)
			exit(EXIT_SUCCESS);

		sid = setsid();
		if (sid < 0) {
			log_message(NAME, L_DEBUG, "setsid() returned error\n");
			exit(EXIT_FAILURE);
		}

		char *directory = "/";

		if ((chdir(directory)) < 0) {
			log_message(NAME, L_DEBUG, "chdir() returned error\n");
			exit(EXIT_FAILURE);
		}
	}
	char *buf = NULL;
	if (asprintf(&buf, "%d", getpid()) != -1) {
		int error = write(fd, buf, strlen(buf));
		if ( error < 0) {
			log_message(NAME, L_DEBUG, "Unable to write the ispappcwmpd pid to /var/run/ispappcwmpd.pid\n");
		}

		free(buf);
	}

	log_message(NAME, L_NOTICE, "entering main loop\n");
	uloop_run();

	ubus_exit();
	uloop_done();

	http_client_exit();
	xml_exit();
	backup_cleanup();
	config_exit();
	cwmp_free_deviceid();
	command_cleanup();

	closelog();
	close(fd);
	if (cwmp->netlink_sock[0] != -1) close(cwmp->netlink_sock[0]);
	if (cwmp->netlink_sock[1] != -1) close(cwmp->netlink_sock[1]);
	free(cwmp);

	log_message(NAME, L_NOTICE, "exiting\n");
	return 0;
}

