import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Pahlawan-Market 📉',
    description: (
      <>
        Dynamic Pricing Engine using Exponential Decay. Prices drop automatically
        as food approaches expiry, ensuring zero waste and affordability.
      </>
    ),
  },
  {
    title: 'Pahlawan-AI 🧠',
    description: (
      <>
        Predictive Analytics that forecasts food waste before it happens.
        Uses historical data and environmental context to alert providers early.
      </>
    ),
  },
  {
    title: 'Pahlawan-Express 🚚',
    description: (
      <>
        Native logistics integration for sub-second courier matching.
        Optimized routes and RT/RW-based group buys to minimize costs.
      </>
    ),
  },
];

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
