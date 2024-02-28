package v1beta2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

// Ec2 store the configuration for services to
// override existing defaults of AWS Services.
type Ec2 struct {
	// ElasticIp is an optional field that can be used to tell the installation process to use
	// Elastic IP address that had been previously created to assign to the resources with Public IPv4
	// address created by installer.
	// +optional
	ElasticIp *Ec2ElasticIp `json:"elasticIp,omitempty"`
}

// Ec2ElasticIp store the configuration for services to
// override existing defaults of AWS Services.
type Ec2ElasticIp struct {
	// PublicIpv4Pool is an optional field that can be used to tell the installation process to use
	// Public IPv4 address that you bring to your AWS account with BYOIP.
	// +optional
	PublicIpv4Pool string `json:"publicIpv4Pool,omitempty"`

	// ElasticIps is an optional field that can be used to tell the installation process to use
	// Elastic IP address that had been previously created to assign to the resources with Public IPv4
	// address created by installer.
	// +optional
	AllocatedIps []string `json:"allocatedIps,omitempty"`
}

// GetOrAllocateAddressesFromBYOIP allocate EIPs from custom configuration.
func (eip *Ec2ElasticIp) getOrAllocateAddressesFromBYOIP(sess ec2iface.EC2API, num int) (eips []string, err error) {

	// TODO: consume user-provided EIP from config (eipConfig.AllocatedIps)
	if len(eip.AllocatedIps) > 0 {
		// TODO validate if the EIP isn't allocated, and consume it to eips.
	}

	// TODO allocate address from BYOIP
	if eip.PublicIpv4Pool == "" {
		return eips, nil
	}
	// TODO allocate EIP from custom BYOIP pool.

	return eips, nil
}

func (eip *Ec2ElasticIp) GetOrAllocateAddresses(sess ec2iface.EC2API, num int, role *string) (eips []string, err error) {

	// 1) Allocate from BYOIP
	// Get custom EIPs from config
	eips, err = eip.getOrAllocateAddressesFromBYOIP(sess, num)
	if err != nil {
		// record.Eventf(s.scope.InfraCluster(), "FailedAllocateBYOIPAddresses", "Failed to allocate EIP from BYOIP: %v", err)
		return nil, errors.Wrap(err, "failed to allocate BYOIP addresses")
	}

	// 2) Lookup unassigned EIPs
	out, err := eip.describeAddresses(sess, "TBD", role)
	if err != nil {
		// record.Eventf(s.scope.InfraCluster(), "FailedDescribeAddresses", "Failed to query addresses for role %q: %v", role, err)
		return nil, errors.Wrap(err, "failed to query addresses")
	}
	if len(eips) < num {
		// TODO fix
		for _, address := range out.Addresses {
			if len(eips) == num {
				break
			}
			if address.AssociationId == nil {
				eips = append(eips, aws.StringValue(address.AllocationId))
			}
		}
	}

	// 3) Allocate EIPs from Amazon-provided IP
	for len(eips) < num {
		ip, err := eip.allocateAddress(sess, role)
		if err != nil {
			return nil, err
		}
		eips = append(eips, ip)
	}

	return eips, nil

}

func (eip *Ec2ElasticIp) allocateAddress(sess ec2iface.EC2API, role *string) (string, error) {
	// tagSpecifications := tags.BuildParamsToTagSpecification(ec2.ResourceTypeElasticIp, s.getEIPTagParams(role))
	out, err := sess.AllocateAddressWithContext(context.TODO(), &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
		// TagSpecifications: []*ec2.TagSpecification{
		// 	tagSpecifications,
		// },
		// PublicIpv4Pool: aws.String("ipv4pool-ec2-09e5e971e86699d07"),
	})
	if err != nil {
		// record.Warnf(s.scope.InfraCluster(), "FailedAllocateEIP", "Failed to allocate Elastic IP for %q: %v", role, err)
		return "", errors.Wrap(err, "failed to allocate Elastic IP")
	}

	return aws.StringValue(out.AllocationId), nil
}

func (eip *Ec2ElasticIp) describeAddresses(sess ec2iface.EC2API, scope string, clusterName *string) (*ec2.DescribeAddressesOutput, error) {
	filters := []*ec2.Filter{{
		Name:   aws.String("tag-key"),
		Values: aws.StringSlice([]string{fmt.Sprintf("kubernetes.io/cluster/%s", *clusterName)}),
	}}
	// if role != "" {
	// 	filters = append(filters, &ec2.Filter{
	// 		Name:   aws.String("tag-key"),
	// 		Values: aws.StringSlice([]string{fmt.Sprintf("kubernetes.io/cluster/%s", role)}),
	// 	})
	// }

	return sess.DescribeAddressesWithContext(context.TODO(), &ec2.DescribeAddressesInput{
		Filters: filters,
	})
}
