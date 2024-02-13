package controller

const (
	// DefaultKyuubiConfFile is the default conf file name of the kyuubi
	DefaultKyuubiConfFile = "kyuubi-defaults.conf"

	// DefaultKyuubiConfDir is the default conf dir of the kyuubi
	DefaultKyuubiConfDir = "/opt/kyuubi/conf"

	// DefaultSparkConfFile is the default conf file name of the spark
	DefaultSparkConfFile = "spark-defaults.conf"

	// DefaultSparkConfDir is the default conf dir of the spark
	DefaultSparkConfDir = "/opt/spark/conf"

	// DefaultHiveSiteFile is the default hive site file name of the metastore
	DefaultHiveSiteFile = "hive-site.xml"

	// DefaultHdfsSiteFile is the default hdfs site file name of the hdfs
	DefaultHdfsSiteFile = "hdfs-site.xml"

	// DefaultCoreSiteFile is the default core site file name of the hdfs
	DefaultCoreSiteFile = "core-site.xml"

	// DefaultNameSuffix is the default name suffix of the resources of the kyuubi
	DefaultNameSuffix = "-kyuubi"

	// DefaultClusterRefsNameSuffix is the default cluster refs config name suffix
	DefaultClusterRefsNameSuffix = "-clusterrefs"

	// DefaultClusterSign is the default cluster sign of the hdfs
	DefaultClusterSign = "kyuubi"

	// FSDefaultFSConfKey is the fs default fs conf key of the hdfs
	FSDefaultFSConfKey = "fs.defaultFS"

	// DFSNameSpacesConfKey is the dfs nameservices conf key of the hdfs
	DFSNameSpacesConfKey = "dfs.nameservices"
)
