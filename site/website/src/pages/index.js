import React from 'react';
import classnames from 'classnames';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './styles.module.css';


const features = [
    {
        title: <>Kubernetes Operator</>,
        imageUrl: 'img/operator-sdk.png',
        description: (
            <>
                NiFiKop will define a new Kubernetes object named NifiCluster which will be used to describe
                and instantiate a Nifi Cluster in Kubernetes
            </>
        ),
    },
    {
        title: <>Open-Source</>,
        imageUrl: 'img/open_source.svg',
        description: (
            <>
                Open source software released under the Apache 2.0 license.
            </>
        ),
    },
    {
        title: <>NiFi Cluster in K8S</>,
        imageUrl: 'img/kubernetes.png',
        description: (
            <>
                NiFiKop is a Kubernetes custom controller which will loop over events on NifiCluster objects and
                reconcile with kubernetes resources needed to create a valid NiFi Cluster deployment.
            </>
        ),
    },
    {
        title: <>Space Scoped</>,
        imageUrl: 'img/namespace.png',
        description: (
            <>
                NiFiKop is listening is a Multi-Namespace scoped operator (not cluster wide), and
                is able to manage several Nifi Clusters within these namespaces.
            </>
        ),
    },

    {
        title: <>User and group management</>,
        imageUrl: 'img/users.png',
        description: (
            <>
                NiFiKop allows you to define users and groups with their access policies using K8s resources.
                This way you can fully automate your NiFi cluster setup using yaml configurations.
            </>
        ),
    },

    {
        title: <>Dataflow lifecycle management</>,
        imageUrl: 'img/dataflow.png',
        description: (
            <>
                NiFiKop allows you to define NiFi registry client, parameter context and datflow using K8s resources.
                This way you can fully automate your Dataflow deployment and let the operator manage is lifecycle.
            </>
        ),
    },
];

function Feature({imageUrl, title, description}) {
    const imgUrl = useBaseUrl(imageUrl);
    return (
        <div className={classnames('col col--4', styles.feature)}>
            {imgUrl && (
                <div className="text--center">
                    <img className={styles.featureImage} src={imgUrl} alt={title} />
                </div>
            )}
            <h3>{title}</h3>
            <p>{description}</p>
        </div>
    );
}

function Home() {
    const context = useDocusaurusContext();
    const {siteConfig: {customFields = {}} = {}} = context;

    return (
        <Layout permalink="/" description={customFields.description}>
            <div className={styles.hero}>
                <div className={styles.heroInner}>
                    <h1 className={styles.heroProjectTagline}>
                        <img
                            alt="NiFiKop"
                            className={styles.heroLogo}
                            src={useBaseUrl('img/nifikop.png')}
                        />
                        Open-Source, Apache <span className={styles.heroProjectKeywords}>NiFi</span>{' '}
                        operator for <span className={styles.heroProjectKeywords}>Kubernetes</span>{' '}
                    </h1>
                    <div className={styles.indexCtas}>
                        <Link
                            className={styles.indexCtasGetStartedButton}
                            to={useBaseUrl('docs/2_setup/1_getting_started')}>
                            Get Started
                        </Link>
                        <span className={styles.indexCtasGitHubButtonWrapper}>
                            <iframe
                                className={styles.indexCtasGitHubButton}
                                src="https://ghbtns.com/github-btn.html?user=Orange-OpenSource&amp;repo=nifikop&amp;type=star&amp;count=true&amp;size=large"
                                width={160}
                                height={30}
                                title="GitHub Stars"
                            />
                        </span>
                    </div>
                </div>
            </div>
            <div className={classnames(styles.announcement, styles.announcementDark)}>
                <div className={styles.announcementInner}>
                    The <span className={styles.heroProjectKeywords}>NiFiKop</span> NiFi Kubernetes operator makes it <span className={styles.heroProjectKeywords}>easy</span> to run Apache NiFi on Kubernetes.
                    Apache NiFI is a free, open-source solution that support powerful and <span className={styles.heroProjectKeywords}>scalable</span> directed graphs of <span className={styles.heroProjectKeywords}>data routing</span>, transformation, and system <span className={styles.heroProjectKeywords}>mediation logic</span>.
                </div>
            </div>
            <div className={styles.section}>
                {features && features.length && (
                    <section className={styles.features}>
                        <div className="container">
                            <div className="row">
                                {features.map((props, idx) => (
                                    <Feature key={idx} {...props} />
                                ))}
                            </div>
                        </div>
                    </section>
                )}
            </div>
        </Layout>
    );
}

export default Home;
