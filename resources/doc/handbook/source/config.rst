.. _config:

:term:`paymentd` configuration
==============================

.. contents::
	:local:

Payment
-------

.. topic:: The Payment section

	::

		"Payment": {
			"PaymentIDEncPrime": 982450871,
			"PaymentIDEncXOR": 123456789
		}

This section contains values related to payments.

*****************
PaymentIDEncPrime
*****************

This is a (large) prime (``int64``), which is used to obfuscate the sequential payment IDs.
This value has to be consistent throughout the whole cluster.

Obfuscation is performed using `Modular multiplicative inverses <http://en.wikipedia.org/wiki/Modular_multiplicative_inverse>`_.

***************
PaymentIDEncXOR
***************

This is an arbitrary ``int64``, which XORs the ModInv of the payment ID.

The pair ``PaymentIDEncPrime`` and ``PaymentIDEncXOR`` is the "secret" which allows
encoding and decoding of payment IDs throughout the cluster.


Database
--------

.. topic:: The Database section

	::

		"Database": {
			"TransactionMaxRetries": 5,
			"MaxOpenConns": 10,
			"MaxIdleConns": 5,
			"Principal": {
				"Write": {
					"mysql": "paymentd@tcp(localhost:3306)/fritzpay_principal?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
				},
				"ReadOnly": null
			},
			"Payment": {
				"Write": {
					"mysql": "paymentd@tcp(localhost:3306)/fritzpay_payment?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
				},
				"ReadOnly": null
			}
		}

The Database section holds values for connecting with the RDBMS (Relational Database
Management System).

:term:`paymentd` operates on two separate databases:

* The principal database.
* The payment database.

Each database can have two modes. A read/write and a read-only mode. A replicated read-only
database can be used to reduce load on the read/write databases.

*********************
TransactionMaxRetries
*********************

The maximum number of retries on database transactions after which a transaction is 
considered failed.

This usually happens when the database cannot get a lock on a row.

************
MaxOpenConns
************

Each database connection (Principal RW, Principal RO, Payment RW, Payment RO) maintains a
connection pool. This is the maximum number of connections for each pool to the
RDBMS and should match the `max_connections <http://dev.mysql.com/doc/refman/5.5/en/server-system-variables.html#sysvar_max_connections>`_ system variable with a reasonable margin
if other processes are connecting to the same RDBMS.

************
MaxIdleConns
************

The connection pools maintain a few open connections to avoid having to reconnect. This
is the maximum number of idle connections allowed.

****
DSNs
****

The connection Data Source Names (DSNs) are described at the `MySQL driver library <https://github.com/go-sql-driver/mysql#dsn-data-source-name>`_.

Important DSN parameters are:

.. table:: Important MySQL DSN parameters

	+--------------------------------+---------------------------------------------------+
	|           Parameter            |                    Explanation                    |
	+================================+===================================================+
	| ``parseTime=true``             | This parameter has to be present so MySQL         |
	|                                | DATETIME fields can be mapped correctly.          |
	+--------------------------------+---------------------------------------------------+
	| ``loc=UTC``                    | This parameter is also required. MySQL uses the   |
	|                                | system timezone, which is almost never desirable. |
	|                                | :term:`paymentd` always uses UTC, therefore       |
	|                                | this parameter will tell MySQL to use UTC for     |
	|                                | DATETIME fields.                                  |
	+--------------------------------+---------------------------------------------------+
	| ``time_zone=%22%2B00%3A00%22`` | ``+00:00`` See `mysql_tz`_                        |
	+--------------------------------+---------------------------------------------------+

.. _mysql_tz: http://dev.mysql.com/doc/refman/5.5/en/server-system-variables.html#sysvar_time_zone

The "Write" DSNs are required. The "ReadOnly" DSNs are optional. If they are ``null``,
only the Read/Write connections will be used.

.. _config_api:

API Service
-----------

.. topic:: The API section

	::

		"API": {
			"Active": true,
			"Service": {
				"Address": ":8080",
				"ReadTimeout": "10s",
				"WriteTimeout": "10s",
				"MaxHeaderBytes": 0
			},
			"Timeout": "5s",
			"ServeAdmin": false,
			"Secure": false,
			"Cookie": {
				"AllowCookieAuth": false,
				"HTTPOnly": true
			},
			"AdminGUIPubWWWDir": "",
			"AuthKeys": []
		}

The API service section holds values for the :ref:`API Server <api_server>`.

******
Active
******

This boolean value indicates whether the server should serve the API service.

***************
Service Address
***************

This is the address the API server will listen on. The default value ``:8080`` listens
on all active interfaces on port ``8080``. If you provide an IP address, the server
will be bound to that IP address.

********************************
Service ReadTimeout/WriteTimeout
********************************

The HTTP timeouts for reading a request and writing a response.

**********************
Service MaxHeaderBytes
**********************

The maximum size of headers. If the default ``0`` is provided, it will be the default
Go ``net.http`` `DefaultMaxHeaderBytes`_ (1 MB at this time).

.. _DefaultMaxHeaderBytes: http://golang.org/pkg/net/http/#pkg-constants

*******
Timeout
*******

A general timeout for all API requests.

.. _config_api_serve_admin:

**********
ServeAdmin
**********

This boolean value indicates whether the API service will also serve administrative
API methods.

******
Secure
******

Whether the API server should be served securely. This affects the secure flags of the
cookies.

While :term:`paymentd` does not support TLS as of now, most installations will run
:term:`paymentd` behind a TLS-enabled proxy. In these cases, this flag should be set
to ``true``.

.. _config_api_cookie_allow_cookie_auth:

**********************
Cookie AllowCookieAuth
**********************

The administrative APIs require a valid ``Authorization`` header and offer means of
obtaining a valid authorization.

When this flag is set to ``true`` obtained authorizations will also set a cookie and
the API endpoints will check for authoriation cookies.

***************
Cookie HTTPOnly
***************

Whether the ``HTTP only`` flag should be set on cookies.

.. _config_api_admin_gui_pub_www_dir:

******************************
Admin GUI Public WWW Directory
******************************

If ``AdminGUIPubWWWDir`` is set to a valid directory, the admin service will also
serve this static directory.

.. _config_api_auth_keys:

********
AuthKeys
********

The API service maintains a list of keys (an array of hex-encoded strings). Those
are used to encrypt the authorization containers. These keys, when shared across 
different services in the whole software stack, allow cross-application
authentication and authorization.

The list is used to roll over new keys. The first key is the preferred key.

.. note::

	Keys will be randomly generated during startup of the daemon, if no keys are
	configured. Those keys must be added to the configuration for persistence.

	Persistence is required to apply the same keys on multiple instances of
	:term:`paymentd` or different applications.

.. _config_www:

Web Server
----------

.. topic:: The Web section

	::

		"Web": {
			"Active": false,
			"URL": "http://localhost:8443",
			"Service": {
				"Address": ":8443",
				"ReadTimeout": "10s",
				"WriteTimeout": "10s",
				"MaxHeaderBytes": 0
			},
			"PubWWWDir": "",
			"TemplateDir": "",
			"Secure": false,
			"Cookie": {
				"HTTPOnly": true
			},
			"AuthKeys": []
		}

The Web service section holds values for the :ref:`Web Server <web_server>`.

******
Active
******

This boolean value indicates whether the server should serve the Web service.

***
URL
***

The URL under which the Web server will be served.

***************
Service Address
***************

This is the address the Web server will listen on. The default value ``:8443`` listens
on all active interfaces on port ``8443``. If you provide an IP address, the server
will be bound to that IP address.

********************************
Service ReadTimeout/WriteTimeout
********************************

The HTTP timeouts for reading a request and writing a response.

**********************
Service MaxHeaderBytes
**********************

The maximum size of headers. If the default ``0`` is provided, it will be the default
Go ``net.http`` `DefaultMaxHeaderBytes`_ (1 MB at this time).

*********
PubWWWDir
*********

The path to the directory where the WWW public files are located. Static HTML/JS/CSS files
should be placed in this directory.

***********
TemplateDir
***********

The path to the directory where the templates are located.

******
Secure
******

Whether the Web server should be served securely. This affects the secure flags of the
cookies.

While :term:`paymentd` does not support TLS as of now, most installations will run
:term:`paymentd` behind a TLS-enabled proxy. In these cases, this flag should be set
to ``true``.

***************
Cookie HTTPOnly
***************

Whether the ``HTTP only`` flag should be set on cookies.

********
AuthKeys
********

The Web service maintains a list of keys (an array of hex-encoded strings). Those
are used to encrypt the payment cookie containers. These keys, when shared across 
different services in the whole software stack, allow cross-application
authentication and authorization.

The list is used to roll over new keys. The first key is the preferred key.

.. note::

	Keys will be randomly generated during startup of the daemon, if no keys are
	configured. Those keys must be added to the configuration for persistence.

	Persistence is required to apply the same keys on multiple instances of
	:term:`paymentd` or different applications.


Provider
--------

.. topic:: The Provider section

	::

		"Provider": {
			"URL": "http://localhost:8443",
			"ProviderTemplateDir": ""
		}

The Provider section holds values for the PSP service.

***
URL
***

The URL under which the provider endpoints will be served.

*******************
ProviderTemplateDir
*******************

The path to the directory which holds the provider templates.

