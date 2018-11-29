package aws

import (
	"sync"
	"time"

	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/cenkalti/backoff"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const dnsNamespace = "openfaas.local"

var once = &sync.Once{}
var namespaceID *string

func deleteServiceRegistration(serviceName string, vpcID string) error {
	namespaceID, err := ensureDNSNamespaceExists(vpcID)
	if err != nil {
		return fmt.Errorf("error ensuring dns namespace existing. %v", err)
	}

	listResults, err := discoveryClient.ListServices(&servicediscovery.ListServicesInput{
		Filters: []*servicediscovery.ServiceFilter{
			{
				Name: aws.String("NAMESPACE_ID"),
				Values: []*string{
					namespaceID,
				},
			},
		},
	})

	serviceID := ""
	for _, item := range listResults.Services {
		if aws.StringValue(item.Name) == serviceName {
			serviceID = aws.StringValue(item.Id)
			break
		}
	}

	if len(serviceID) == 0 {
		return nil // nothing to do
	}

	log.Infof("Listing service instances for %s", serviceID)
	instances, err := discoveryClient.ListInstances(&servicediscovery.ListInstancesInput{
		ServiceId: aws.String(serviceID),
	})
	if err != nil {
		return fmt.Errorf("error listing service discovery instance for %s with id %s. %v",
			serviceName, serviceID, err)
	}

	for _, v := range instances.Instances {
		log.Infof("De-registering instance %s for service %s", aws.StringValue(v.Id), serviceID)

		_, err = discoveryClient.DeregisterInstance(&servicediscovery.DeregisterInstanceInput{
			ServiceId:  aws.String(serviceID),
			InstanceId: v.Id,
		})

		if err != nil {
			return fmt.Errorf("error de-registering service discovery instance id %s, for service %s with id %s. %v",
				aws.StringValue(v.Id), serviceName, serviceID, err)
		}
	}

	eb := backoff.NewExponentialBackOff()
	eb.MaxElapsedTime = time.Second * 30

	err = backoff.Retry(func() error {
		_, err := discoveryClient.DeleteService(&servicediscovery.DeleteServiceInput{
			Id: aws.String(serviceID),
		})

		if err != nil {
			log.Errorf("error deleting service discovery service %s with id %s. %v. don't worry we are going to back off and retry...", serviceName, serviceID, err)
		} else {
			log.Infof("yay we deleted service %s with id %s.", serviceName, serviceID)
		}

		return err
	}, eb)

	if err != nil {
		return fmt.Errorf("error deleting service discovery service %s with id %s. %v", serviceName, serviceID, err)
	}

	return nil
}

func ensureServiceRegistrationExists(serviceName string, vpcID string) (string, error) {

	namespaceID, err := ensureDNSNamespaceExists(vpcID)
	if err != nil {
		log.Errorln("error ensuring dns namespace existing. ", err)
		return "", err
	}

	listResults, err := discoveryClient.ListServices(&servicediscovery.ListServicesInput{
		Filters: []*servicediscovery.ServiceFilter{
			{
				Name: aws.String("NAMESPACE_ID"),
				Values: []*string{
					namespaceID,
				},
			},
		},
	})

	if err != nil {
		log.Errorln("error listing route 53 auto-naming services. ", err)
		return "", err
	}

	serviceArn := ""
	for _, item := range listResults.Services {
		if aws.StringValue(item.Name) == serviceName {
			serviceArn = aws.StringValue(item.Arn)
			break
		}
	}

	if serviceArn == "" {
		requestID := uuid.NewV4()
		createResult, err := discoveryClient.CreateService(&servicediscovery.CreateServiceInput{
			Name:             aws.String(serviceName),
			CreatorRequestId: aws.String(requestID.String()),
			Description:      aws.String(fmt.Sprintf("Openfaas auto-naming service for %s", serviceName)),
			DnsConfig: &servicediscovery.DnsConfig{
				NamespaceId: namespaceID,
				DnsRecords: []*servicediscovery.DnsRecord{
					{
						Type: aws.String("A"),
						TTL:  aws.Int64(10),
					},
				},
			},
			HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(1),
			},
		})

		if err != nil {
			log.Errorln(fmt.Sprintf("error creating route 53 auto-naming services for %s. ", serviceName), err)
			return "", err
		}

		serviceArn = aws.StringValue(createResult.Service.Arn)
	}

	return serviceArn, nil
}

func ensureDNSNamespaceExists(vpcID string) (id *string, err error) {
	once.Do(func() {
		var found bool

		id, found, err = findNamespace()
		if err != nil {
			log.Errorln("error finding private dns name. ", err)
			return
		}

		if !found {
			requestID := uuid.NewV4()
			_, err = discoveryClient.CreatePrivateDnsNamespace(&servicediscovery.CreatePrivateDnsNamespaceInput{
				Name:             aws.String(dnsNamespace),
				CreatorRequestId: aws.String(requestID.String()),
				Description:      aws.String("Openfaas private DNS namespace"),
				Vpc:              aws.String(vpcID),
			})

			if err != nil {
				log.Errorln("error creating private dns name. ", err)
				return
			}

			id, found, err = findNamespace()
			if err != nil {
				log.Errorln("error finding private dns name. ", err)
				return
			}

			if !found {
				log.Errorln("could not find private dns after creating it")
				err = errors.New("could not find private dns after creating it")
			}
		}

		namespaceID = id
	})

	return namespaceID, err
}

func findNamespace() (*string, bool, error) {
	var listResult *servicediscovery.ListNamespacesOutput
	listResult, err := discoveryClient.ListNamespaces(&servicediscovery.ListNamespacesInput{})
	if err != nil {
		log.Errorln("error listing namespaces. ", err)
		return nil, false, err
	}

	found := false
	var id *string
	for _, item := range listResult.Namespaces {
		if aws.StringValue(item.Name) == dnsNamespace {
			id = item.Id
			found = true
			break
		}
	}

	return id, found, err
}
