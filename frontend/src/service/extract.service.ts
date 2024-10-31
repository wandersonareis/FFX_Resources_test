import { Injectable } from '@angular/core';
import { Extract } from '../../wailsjs/go/services/ExtractService';
import { interactions } from '../../wailsjs/go/models';

@Injectable({
  providedIn: 'root'
})
export class ExtractService {
  async extraction(fileInfo: interactions.GameDataInfo) {
    Extract(fileInfo);
  }
}
