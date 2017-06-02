#!/bin/bash

kubeMaster=http://10.10.174.116:8080
./reflectorTest --v 4 \
                --masterUrl $kubeMaster \
                --alsologtostderr

#./reflectorTest --v 4 \
#                --kubeconfig ./configs/aws.kubeconfig.yaml \
#                --alsologtostderr
